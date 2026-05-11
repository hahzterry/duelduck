'use client';

import axios from 'axios';
import { getSession } from 'next-auth/react';

import { BASE_URL } from '~constants/api';

const api = axios.create({ baseURL: BASE_URL });

api.interceptors.request.use(async (config) => {
  const session = await getSession();
  const accessToken = session?.accessToken;

  if (accessToken) {
    config.headers = config.headers ?? {};
    config.headers.Authorization = `Bearer ${accessToken}`;
  }

  return config;
});

export default api;
