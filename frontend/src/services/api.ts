import axios from "axios";
import { getAccessToken } from "../auth/keycloak";

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? "http://localhost:8080",
  withCredentials: true
});

api.interceptors.request.use(async (requestConfig) => {
  const token = await getAccessToken();
  if (token) {
    requestConfig.headers = requestConfig.headers ?? {};
    requestConfig.headers.Authorization = `Bearer ${token}`;
  }
  return requestConfig;
});
