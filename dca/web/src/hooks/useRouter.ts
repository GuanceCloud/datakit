import { useNavigate } from "react-router-dom";

export default function useRouter () {
  const navigate = useNavigate()
  return {navigate}
}