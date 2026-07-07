import { ProjectSettings } from "@/components/project-settings";
export default async function ProjectPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = await params;
  return <ProjectSettings projectId={projectId} />;
}
