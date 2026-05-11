import type { CustomDuelType } from '~types/createDuel';

const DB_NAME = 'duelduck-create-duel';
const DB_VERSION = 1;
const STORE_NAME = 'drafts';
const LOCAL_STORAGE_PREFIX = 'duelduck-create-duel-draft:';
const ANONYMOUS_OWNER = 'anonymous';

export const PENDING_CREATE_DUEL_REDIRECT_KEY =
  'duelduck-pending-create-duel-redirect';

export interface CreateDuelDraft {
  duel: CustomDuelType;
  openedItems: string[];
  isCheckedFields: Record<string, boolean>;
  tokenImage: string | null;
  savedAt: number;
}

type DraftOwner = string | null;

const ownerKey = (userId: DraftOwner): string =>
  userId ? userId : ANONYMOUS_OWNER;

const localStorageKey = (userId: DraftOwner): string =>
  `${LOCAL_STORAGE_PREFIX}${ownerKey(userId)}`;

const writeLocalStorageDraft = (draft: CreateDuelDraft, userId: DraftOwner) => {
  if (typeof window === 'undefined' || !window.localStorage) return;

  try {
    window.localStorage.setItem(localStorageKey(userId), JSON.stringify(draft));
  } catch {
    // localStorage may be unavailable (Safari private mode, quota); IDB is primary.
  }
};

const readLocalStorageDraft = (userId: DraftOwner): CreateDuelDraft | null => {
  if (typeof window === 'undefined' || !window.localStorage) return null;

  try {
    const raw = window.localStorage.getItem(localStorageKey(userId));

    return raw ? (JSON.parse(raw) as CreateDuelDraft) : null;
  } catch {
    return null;
  }
};

const clearLocalStorageDraft = (userId: DraftOwner) => {
  if (typeof window === 'undefined' || !window.localStorage) return;

  try {
    window.localStorage.removeItem(localStorageKey(userId));
  } catch {
    // ignore
  }
};

const openDraftDb = (): Promise<IDBDatabase> =>
  new Promise((resolve, reject) => {
    if (typeof window === 'undefined' || !window.indexedDB) {
      reject(new Error('IndexedDB is not available'));

      return;
    }

    const request = window.indexedDB.open(DB_NAME, DB_VERSION);

    request.onupgradeneeded = () => {
      const db = request.result;

      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME);
      }
    };

    request.onsuccess = () => resolve(request.result);
    request.onerror = () =>
      reject(request.error || new Error('Failed to open draft database'));
  });

const withDraftStore = async <T>(
  mode: IDBTransactionMode,
  callback: (store: IDBObjectStore) => IDBRequest<T>,
): Promise<T> => {
  const db = await openDraftDb();

  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, mode);
    const store = tx.objectStore(STORE_NAME);
    const request = callback(store);

    request.onsuccess = () => resolve(request.result);
    request.onerror = () =>
      reject(request.error || new Error('Draft database request failed'));
    tx.oncomplete = () => db.close();
    tx.onerror = () => {
      db.close();
      reject(tx.error || new Error('Draft database transaction failed'));
    };
  });
};

export const saveCreateDuelDraft = async (
  draft: CreateDuelDraft,
  userId: DraftOwner,
): Promise<void> => {
  const key = ownerKey(userId);

  writeLocalStorageDraft(draft, userId);

  try {
    await withDraftStore('readwrite', (store) => store.put(draft, key));
  } catch {
    // IDB may be unavailable; localStorage mirror is enough to survive auth flows.
  }
};

export const loadCreateDuelDraft = async (
  userId: DraftOwner,
): Promise<CreateDuelDraft | null> => {
  const key = ownerKey(userId);

  try {
    const draft = await withDraftStore<CreateDuelDraft | undefined>(
      'readonly',
      (store) => store.get(key),
    );

    if (draft) return draft;
  } catch {
    // fall through to localStorage
  }

  return readLocalStorageDraft(userId);
};

export const clearCreateDuelDraft = async (
  userId: DraftOwner,
): Promise<void> => {
  const key = ownerKey(userId);

  clearLocalStorageDraft(userId);

  try {
    await withDraftStore('readwrite', (store) => store.delete(key));
  } catch {
    // Best-effort cleanup; publishing should not fail because browser storage did.
  }
};

export const claimAnonymousCreateDuelDraft = async (
  userId: string,
): Promise<void> => {
  if (!userId) return;

  const anonymousDraft = await loadCreateDuelDraft(null);

  if (!anonymousDraft) return;

  await saveCreateDuelDraft(anonymousDraft, userId);
  await clearCreateDuelDraft(null);
};
