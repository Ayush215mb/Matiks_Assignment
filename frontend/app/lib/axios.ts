import axios from "axios";

const resolvedBaseURL = "http://localhost:8080";

export const apiClient = axios.create({
  baseURL: resolvedBaseURL,
  timeout: 15000,
  headers: {
    "Content-Type": "application/json",
  },
});

// Centralized error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error("API Error:", error.response?.data || error.message);

    return Promise.reject(error);
  },
);

export default apiClient;
