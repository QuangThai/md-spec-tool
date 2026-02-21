import { useWorkbenchDataFlows } from "@/hooks/useWorkbenchDataFlows";
import { useWorkbenchActions } from "@/hooks/useWorkbenchActions";
import type { WorkbenchDomainModel } from "@/hooks/useWorkbenchContracts";

export function useWorkbenchDomain(): WorkbenchDomainModel {
  const data = useWorkbenchDataFlows();
  const actions = useWorkbenchActions(data);

  return {
    ...data,
    ...actions,
  };
}
