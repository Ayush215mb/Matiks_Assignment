import { apiClient } from "../lib/axios";

export type Player = {
    "rank": string
    "username":string
    "rating": number
};

export type PlayerWithRank = Player & { rank: number };

export type LeaderboardResponse = {
    entries: Player[];
    total?: number;
};

export type SearchResponse = {
    entries: PlayerWithRank[];
};

export const getLeaderboard = async (): Promise<Player[]> => {
    try {
        const response = await apiClient.get<LeaderboardResponse>("/api/leaderboard?page=1&limit=50");
        const Playerdata= response.data.entries
        return Playerdata || response.data as unknown as Player[];
    } catch (err: unknown) {
        console.error("Error fetching leaderboard:", err);
        throw err;
    }
};

export const searchUser = async (username: string): Promise<PlayerWithRank[]> => {
    try {
        const response = await apiClient.get<SearchResponse>(`/api/search?q=${encodeURIComponent(username)}`);

        return response.data.results
    } catch (error) {
        console.error("Error searching user:", error);
        throw error;
    }
};

export const seedData = async (count: number): Promise<unknown> => {
    try {
        const response = await apiClient.post("/api/seed", { count });
        return response.data;
    } catch (error) {
        console.error("Error seeding data:", error);
        throw error;
    }
};


