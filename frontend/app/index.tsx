import { StatusBar } from "expo-status-bar";
import { useCallback, useEffect, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  Pressable,
  RefreshControl,
  Text,
  TextInput,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import {
  getLeaderboard,
  searchUser,
  seedData,
  type PlayerWithRank,
} from "./services/api";

export default function Index() {
  const [leaderboard, setLeaderboard] = useState<PlayerWithRank[]>([]);
  const [searchText, setSearchText] = useState("");
  const [searchResults, setSearchResults] = useState<PlayerWithRank[] | null>(
    null,
  );
  const [refreshing, setRefreshing] = useState(false);
  const [loading, setLoading] = useState(true);
  const [searchLoading, setSearchLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [totalUsers, setTotalUsers] = useState(0);
  const [loadingMore, setLoadingMore] = useState(false);
  const limit = 50;

  const fetchLeaderboard = useCallback(
    async (pageNum: number = 1, append: boolean = false) => {
      try {
        setError(null);
        setLoadingMore(append);

        const response = await getLeaderboard(pageNum, limit);

        if (append) {
          setLeaderboard((prev) => [...prev, ...response.entries]);
        } else {
          setLeaderboard(response.entries);
        }

        setPage(response.page);
        setHasMore(response.has_more);
        setTotalUsers(response.total_users);
      } catch (err: any) {
        console.error("Failed to fetch leaderboard:", err);
        const isNetworkError =
          err?.message?.includes("Network Error") ||
          err?.code === "ERR_NETWORK";
        if (isNetworkError) {
          setError(
            "Cannot connect to server. Make sure the backend is running on port 8080.",
          );
        } else {
          setError("Failed to load leaderboard. Please try again.");
        }
      } finally {
        setLoading(false);
        setRefreshing(false);
        setLoadingMore(false);
      }
    },
    [limit],
  );

  useEffect(() => {
    const initializeData = async () => {
      try {
        setLoading(true);
        // Seed data first, then fetch leaderboard
        await seedData(10000);
        await fetchLeaderboard(1, false);
      } catch (err) {
        console.error("Failed to initialize:", err);
        setError("Failed to initialize. Please try again.");
        setLoading(false);
      }
    };

    initializeData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleRefresh = useCallback(() => {
    setRefreshing(true);
    setPage(1);
    fetchLeaderboard(1, false);
  }, [fetchLeaderboard]);

  const handleLoadMore = useCallback(() => {
    if (hasMore && !loadingMore && !refreshing) {
      fetchLeaderboard(page + 1, true);
    }
  }, [hasMore, loadingMore, refreshing, page, fetchLeaderboard]);

  const handleSearch = useCallback(async () => {
    const query = searchText.trim();
    if (!query) {
      setSearchResults(null);
      return;
    }

    setSearchLoading(true);
    try {
      const results = await searchUser(query);
      setSearchResults(results);
      setError(null);
    } catch (err: any) {
      console.error("Failed to search user:", err);
      const isNetworkError =
        err?.message?.includes("Network Error") || err?.code === "ERR_NETWORK";
      setSearchResults([]);
      if (isNetworkError) {
        setError(
          "Cannot connect to server. Make sure the backend is running on port 8080.",
        );
      } else {
        setError("Failed to search. Please try again.");
      }
    } finally {
      setSearchLoading(false);
    }
  }, [searchText]);

  const renderRow = ({ item }: { item: PlayerWithRank }) => (
    <View className="flex-row items-center justify-between rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
      <Text className="w-12 text-center text-base font-semibold text-white">
        #{item.rank}
      </Text>
      <Text className="flex-1 text-base font-semibold text-white">
        {item.username}
      </Text>
      <Text className="text-base font-semibold text-emerald-300">
        {item.rating}
      </Text>
    </View>
  );

  const renderFooter = () => {
    if (!loadingMore) return null;
    return (
      <View className="py-4">
        <ActivityIndicator size="small" color="#34d399" />
      </View>
    );
  };

  const renderSearchResult = () => {
    if (!searchText.trim()) {
      return (
        <Text className="text-sm text-slate-300">
          Type a username to search for their live global rank.
        </Text>
      );
    }

    if (searchLoading) {
      return (
        <View className="flex-row items-center justify-center py-4">
          <ActivityIndicator size="small" color="#34d399" />
          <Text className="ml-2 text-sm text-slate-300">Searching...</Text>
        </View>
      );
    }

    if (!searchResults || searchResults.length === 0) {
      return (
        <Text className="text-sm font-semibold text-rose-300">
          No users found for &quot;{searchText.trim()}&quot;.
        </Text>
      );
    }

    return (
      <View className="gap-2">
        <View className="flex-row items-center justify-between">
          <Text className="text-sm font-semibold uppercase tracking-wide text-emerald-200">
            Results · {searchResults.length}
          </Text>
          <Text className="text-xs text-slate-400">Live global ranks</Text>
        </View>
        <FlatList
          data={searchResults}
          keyExtractor={(item) => item.username}
          renderItem={({ item }) => (
            <View className="flex-row items-center justify-between rounded-2xl border border-emerald-400/40 bg-emerald-500/10 px-4 py-3">
              <View className="w-16">
                <Text className="text-xs uppercase tracking-wide text-emerald-200">
                  Rank
                </Text>
                <Text className="text-lg font-bold text-white">
                  #{item.rank}
                </Text>
              </View>
              <View className="flex-1 pl-3">
                <Text className="text-sm font-semibold text-white">
                  {item.username}
                </Text>
              </View>
              <Text className="text-base font-bold text-emerald-200">
                {item.rating}
              </Text>
            </View>
          )}
          contentContainerStyle={{ gap: 8 }}
          style={{ maxHeight: 320 }}
          nestedScrollEnabled
          showsVerticalScrollIndicator={false}
        />
      </View>
    );
  };

  if (loading) {
    return (
      <SafeAreaView className="flex-1 bg-slate-950">
        <StatusBar style="light" />
        <View className="flex-1 items-center justify-center">
          <ActivityIndicator size="large" color="#34d399" />
          <Text className="mt-4 text-base text-slate-300">
            Loading leaderboard...
          </Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView className="flex-1 bg-slate-950">
      <StatusBar style="light" />
      <View className="flex-1 w-full self-center px-4 py-6 md:max-w-5xl">
        <View className="gap-2 pb-4">
          <Text className="text-3xl font-extrabold text-white">
            Live Leaderboard
          </Text>
          <Text className="text-base text-slate-300">
            Real-time rankings updated from the server.
          </Text>
        </View>

        {error && (
          <View className="mb-4 rounded-xl border border-rose-500/50 bg-rose-500/10 px-4 py-3">
            <Text className="text-sm font-semibold text-rose-300">{error}</Text>
          </View>
        )}

        <View className="gap-3 rounded-2xl border border-white/10 bg-white/5 p-4">
          <Text className="text-lg font-semibold text-white">
            Search player
          </Text>
          <View className="flex-row items-center gap-3">
            <TextInput
              placeholder="Search by username"
              placeholderTextColor="#cbd5e1"
              value={searchText}
              onChangeText={setSearchText}
              className="flex-1 rounded-xl border border-white/10 bg-white/10 px-4 py-3 text-base text-white"
              returnKeyType="search"
              onSubmitEditing={handleSearch}
              autoCapitalize="none"
              autoCorrect={false}
            />
            <Pressable
              onPress={handleSearch}
              disabled={searchLoading}
              className="rounded-xl bg-emerald-500 px-4 py-3 disabled:opacity-50"
              accessibilityRole="button"
            >
              <Text className="text-base font-semibold text-slate-950">
                Search
              </Text>
            </Pressable>
          </View>
          {renderSearchResult()}
        </View>

        <View className="mt-6 flex-1">
          <View className="mb-3 flex-row items-center justify-between">
            <Text className="text-lg font-semibold text-white">
              Top players
            </Text>
            <Text className="text-sm text-slate-400">
              Showing {leaderboard.length} of {totalUsers} players
            </Text>
          </View>

          {leaderboard.length === 0 ? (
            <View className="flex-1 items-center justify-center py-12">
              <Text className="text-base text-slate-400">
                No players found. Pull down to refresh.
              </Text>
            </View>
          ) : (
            <FlatList
              data={leaderboard}
              keyExtractor={(item) => item.username}
              renderItem={renderRow}
              contentContainerStyle={{ gap: 8, paddingBottom: 24 }}
              keyboardShouldPersistTaps="handled"
              refreshControl={
                <RefreshControl
                  refreshing={refreshing}
                  onRefresh={handleRefresh}
                  tintColor="#34d399"
                />
              }
              onEndReached={handleLoadMore}
              onEndReachedThreshold={0.5}
              ListFooterComponent={renderFooter}
              initialNumToRender={20}
              removeClippedSubviews
            />
          )}
        </View>
      </View>
    </SafeAreaView>
  );
}
