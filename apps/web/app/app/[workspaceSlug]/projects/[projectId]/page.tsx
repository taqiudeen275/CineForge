import { ProjectSettings } from "@/components/project-settings";
export default async function ProjectPage({ params }: { params: Promise<{ projectId: string; workspaceSlug:string }> }) {
  const { projectId,workspaceSlug } = await params;
  return <ProjectSettings projectId={projectId} workspaceSlug={workspaceSlug} />;
}
