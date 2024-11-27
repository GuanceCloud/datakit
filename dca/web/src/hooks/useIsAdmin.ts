import { useCallback, useEffect, useState } from "react";
import { useLazyGetAccountPermissionsQuery } from "src/store/consoleApi";
import { IAccountPermission } from "src/store/type";

export function useIsAdmin(): boolean {
  const [isAdmin, setIsAdmin] = useState(false);
  const [getAccountPermissions, { data: accountPermissions }] = useLazyGetAccountPermissionsQuery()

  const isAdminByPermission = useCallback(
    (permission: IAccountPermission): boolean => {
      if (!permission) {
        return false
      }

      if (permission?.roles.length > 0 &&
        (permission.roles.includes("owner") || permission.roles.includes("wsAdmin"))) {
        return true
      }

      return false
    }
    , [])

  useEffect(() => {
    if (accountPermissions?.code === 200) {
      setIsAdmin(isAdminByPermission(accountPermissions?.content))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [accountPermissions])


  useEffect(() => {
    getAccountPermissions()

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return isAdmin
}