import { FIREBASE_CONFIG } from '~constants/api';
import { getTrackingTokens } from '~utils/cookieManager';

// Lazy Firebase initialization
let firebaseApp: any = null;
let authInstance: any = null;

const getFirebaseAuth = async () => {
  if (!authInstance) {
    const { initializeApp } = await import('firebase/app');
    const { getAuth, GoogleAuthProvider } = await import('firebase/auth');

    firebaseApp = initializeApp(FIREBASE_CONFIG);
    const auth = getAuth(firebaseApp);
    const provider = new GoogleAuthProvider();

    provider.setCustomParameters({
      prompt: 'select_account ',
    });
    authInstance = { auth, provider };
  }

  return authInstance;
};

// Lazy auth getter for components that need auth object
export const getAuth = async () => {
  const { auth } = await getFirebaseAuth();

  return auth;
};

export const signInWithGooglePopup = async () => {
  localStorage.removeItem('wallet');
  localStorage.removeItem('withWallet');
  localStorage.removeItem('chain');
  localStorage.removeItem('promoShowed');

  const { auth, provider } = await getFirebaseAuth();
  const { signInWithPopup } = await import('firebase/auth');
  const { signIn } = await import('next-auth/react');

  const r = await signInWithPopup(auth, provider);
  const firebaseIdToken = await r.user.getIdToken();

  const body = getTrackingTokens();

  try {
    const res = await signIn('google-firebase', {
      redirect: false,
      firebase_id_token: firebaseIdToken,
      ...body,
    });

    if (res?.ok) {
      localStorage.removeItem('link_token');

      return true;
    }
  } catch (e) {
    console.error('error login', e);
  }

  return false;
};
