import { userApi, type UserProfileResponse } from "@/lib/api/user"
import { useAsync } from "./use-async"

export function useAuth() {
  const { data: user, loading } = useAsync(async () => {
    return await userApi.getProfile()
  }, null as unknown as UserProfileResponse)

  return {
    user,
    loading,
    isAuthenticated: !!user?.id
  }
}
