import { StatusBar } from "expo-status-bar";
import { useEffect, useMemo, useState } from "react";
import {
  FlatList,
  Pressable,
  RefreshControl,
  Text,
  TextInput,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

type Player = {
  id: string;
  username: string;
  rating: number;
};

type PlayerWithRank = Player & { rank: number };

const LEADERBOARD_SIZE = 1000;
const MIN_RATING = 100;
const MAX_RATING = 5000;

const createInitialData = (): Player[] =>
  Array.from({ length: LEADERBOARD_SIZE }, (_, index) => ({
    id: String(index + 1),
    username: `player_${String(index + 1).padStart(3, "0")}`,
    rating: MIN_RATING + Math.floor(Math.random() * (MAX_RATING - MIN_RATING)),
  }));

const getUpdatedRatings = (players: Player[]): Player[] =>
  players.map((player) => {
    const swing = Math.floor(Math.random() * 120) - 60; // -60 to +60
    const next = player.rating + swing;
    const clamped = Math.min(MAX_RATING, Math.max(MIN_RATING, next));
    return { ...player, rating: clamped };
  });

export default function Index() {
  const [leaderboard, setLeaderboard] = useState<Player[]>(createInitialData);
  const [searchText, setSearchText] = useState("");
  const [searchResults, setSearchResults] = useState<PlayerWithRank[] | null>(
    null,
  );
  const [refreshing, setRefreshing] = useState(false);

  // Keep the board fresh to mimic live rating changes.
  useEffect(() => {
    const intervalId = setInterval(() => {
      setLeaderboard((prev) => getUpdatedRatings(prev));
    }, 4000);

    return () => clearInterval(intervalId);
  }, []);

  const sortedLeaderboard = useMemo(
    () => [...leaderboard].sort((a, b) => b.rating - a.rating),
    [leaderboard],
  );

  const leaderboardWithRank = useMemo(
    () => {
      // Dense ranking: same rating -> same rank; next distinct rating increments by 1.
      const ranked: PlayerWithRank[] = [];
      let currentRank = 0;
      let lastRating: number | null = null;

      sortedLeaderboard.forEach((player, index) => {
        if (player.rating !== lastRating) {
          currentRank = currentRank + 1;
          lastRating = player.rating;
        }

        ranked.push({ ...player, rank: currentRank });
      });

      return ranked;
    },
    [sortedLeaderboard],
  );

  const handleRefresh = () => {
    setRefreshing(true);
    setLeaderboard((prev) => getUpdatedRatings(prev));
    setRefreshing(false);
  };

  const handleSearch = () => {
    const query = searchText.trim().toLowerCase();
    if (!query) {
      setSearchResults(null);
      return;
    }

    const matches = leaderboardWithRank.filter((player) =>
      player.username.toLowerCase().includes(query),
    );

    setSearchResults(matches);
  };

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

  const renderSearchResult = () => {
    if (!searchText.trim()) {
      return (
        <Text className="text-sm text-slate-300">
          Type a username (e.g. player_042) to see their live rank.
        </Text>
      );
    }

    if (!searchResults || searchResults.length === 0) {
      return (
        <Text className="text-sm font-semibold text-rose-300">
          No live rank found for {searchText.trim()}.
        </Text>
      );
    }

    return (
      <View className="gap-2">
        <View className="flex-row items-center justify-between">
          <Text className="text-sm font-semibold uppercase tracking-wide text-emerald-200">
            Results · {searchResults.length}
          </Text>
          <Text className="text-xs text-slate-400">
            Live ranks update every few seconds
          </Text>
        </View>
        <FlatList
          data={searchResults}
          keyExtractor={(item) => item.id}
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

  return (
    <SafeAreaView className="flex-1 bg-slate-950">
      <StatusBar style="light" />
      <View className="flex-1 w-full self-center px-4 py-6 md:max-w-5xl">
        <View className="gap-2 pb-4">
          <Text className="text-3xl font-extrabold text-white">
            Live Leaderboard
          </Text>
          <Text className="text-base text-slate-300">
            Updated every few seconds to mimic thousands of active players.
          </Text>
        </View>

        <View className="gap-3 rounded-2xl border border-white/10 bg-white/5 p-4">
          <Text className="text-lg font-semibold text-white">Search player</Text>
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
              className="rounded-xl bg-emerald-500 px-4 py-3"
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
            <Text className="text-lg font-semibold text-white">Top players</Text>
            <Text className="text-sm text-slate-400">
              Showing {leaderboardWithRank.length} players
            </Text>
          </View>

          <FlatList
            data={leaderboardWithRank}
            keyExtractor={(item) => item.id}
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
            initialNumToRender={20}
            removeClippedSubviews
          />
        </View>
      </View>
    </SafeAreaView>
  );
}
