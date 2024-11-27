export function getToken (): string {
  const tokenString = sessionStorage.getItem('token');
  return tokenString || ""
}

export function delToken() {
  sessionStorage.removeItem('token')
}

export function saveToken(token: string){
  sessionStorage.setItem('token', token)
}

export function getWorkspaceID (): string {
  const tokenString = sessionStorage.getItem('workspaceID');
  return tokenString || ""
}

export function delWorkspaceID() {
  sessionStorage.removeItem('workspaceID')
}

export function saveWorkspaceID(id: string){
  sessionStorage.setItem('workspaceID', id)
}
