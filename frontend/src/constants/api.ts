export const BASE_URL = process.env.NEXT_PUBLIC_DUEL_DUCK_API;

export const IS_STAGE = BASE_URL?.includes('stage');

export const LINK_DD = `https://${IS_STAGE ? 'stage.' : ''}duelduck.com`;

export const SHARE_DUEL_URL = LINK_DD;

export const DUEL_DUCK_API = process.env.NEXT_PUBLIC_DUEL_DUCK_API;

export const RPC_URL = process.env.NEXT_PUBLIC_RPC_URL as string;

export const FIREBASE_CONFIG = {
  apiKey: 'AIzaSyD4VOeOUD1klrPlC8mK5xcJ7JNB_-5rk88',
  authDomain: 'duel-duck.firebaseapp.com',
  projectId: 'duel-duck',
  storageBucket: 'duel-duck.appspot.com',
  messagingSenderId: '170995853083',
  appId: '1:170995853083:web:39f20634e84f8d255295be',
  measurementId: 'G-R3M7PFX7X7',
};
