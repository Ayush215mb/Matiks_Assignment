import { apiClient } from "../lib/axios";

export type Player = {
  rank: number;
  username: string;
  rating: number;
};

export type PlayerWithRank = Player & { rank: number };

export type LeaderboardResponse = {
  entries: Player[];
  page: number;
  limit: number;
  total_users: number;
  has_more: boolean;
};

export type SearchResponse = {
  count: number;
  results: PlayerWithRank[];
};

export const getLeaderboard = async (
  page: number = 1,
  limit: number = 50,
): Promise<LeaderboardResponse> => {
  try {
    const response = await apiClient.get<LeaderboardResponse>(
      `/api/leaderboard?page=${page}&limit=${limit}`,
    );
    return response.data;
  } catch (err: unknown) {
    console.error("Error fetching leaderboard:", err);
    throw err;
  }
};

export const searchUser = async (
  username: string,
): Promise<PlayerWithRank[]> => {
  try {
    const response = await apiClient.get<SearchResponse>(
      `/api/search?q=${encodeURIComponent(username)}`,
    );
    return response.data.results || [];
  } catch (error) {
    console.error("Error searching user:", error);
    throw error;
  }
};

export const seedData = async (count: number): Promise<void> => {
  try {
    await apiClient.post("/api/seed", { count });
  } catch (error) {
    console.error("Error seeding data:", error);
    throw error;
  }
};
