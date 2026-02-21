import { useWorkbenchDomain } from "@/hooks/useWorkbenchDomain";
import { useWorkbenchPresentationProps } from "@/hooks/useWorkbenchPresentationProps";

export function useWorkbenchController() {
  const domain = useWorkbenchDomain();
  return useWorkbenchPresentationProps(domain);
}
