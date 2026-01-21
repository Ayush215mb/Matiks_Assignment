import axios from "axios";
import { Platform } from "react-native";

// Get base URL based on platform
// Android emulator uses 10.0.2.2 to access host machine's localhost
// iOS simulator and web can use localhost
// For physical devices, use your computer's IP address (e.g., http://192.168.x.x:8080)
const getBaseURL = () => {
    // Allow override via environment variable
    if (process.env.EXPO_PUBLIC_API_URL) {
        return process.env.EXPO_PUBLIC_API_URL;
    }
    
    if (Platform.OS === "android") {
        // Android emulator
        return "http://10.0.2.2:8080";
    }
    
    // iOS simulator and web
    return "http://localhost:8080";
};

const resolvedBaseURL = getBaseURL();

console.log(`API Base URL: ${resolvedBaseURL} (Platform: ${Platform.OS})`);

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
        // Log error for debugging
        console.error("API Error:", error.response?.data || error.message);
        
        // Return the error as-is so components can handle it
        return Promise.reject(error);
    }
);

export default apiClient;