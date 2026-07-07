import type { components } from "@cineforge/api-client";

export type ProjectSettings = components["schemas"]["ProjectSettings"];

export interface ProjectSettingsFormValues {
  name: string;
  aspect: string;
  fps: string;
  language: string;
  rating: string;
  style: string;
  quality: string;
}

export function serializeProjectSettings(
  current: ProjectSettings,
  values: ProjectSettingsFormValues,
): ProjectSettings {
  const [aspectWidth = 16, aspectHeight = 9] = values.aspect.split(":").map(Number);
  const [frameRateNumerator = 24, frameRateDenominator = 1] = values.fps.split("/").map(Number);

  return {
    ...current,
    name: values.name,
    aspectWidth,
    aspectHeight,
    frameRateNumerator,
    frameRateDenominator,
    audioLanguage: values.language,
    rating: values.rating as ProjectSettings["rating"],
    styleDirection: values.style,
    qualityPolicy: values.quality as ProjectSettings["qualityPolicy"],
  };
}

export function formValuesFromFormData(form: FormData): ProjectSettingsFormValues {
  return {
    name: String(form.get("name") ?? ""),
    aspect: String(form.get("aspect") ?? "16:9"),
    fps: String(form.get("fps") ?? "24/1"),
    language: String(form.get("language") ?? "en"),
    rating: String(form.get("rating") ?? "moderate"),
    style: String(form.get("style") ?? ""),
    quality: String(form.get("quality") ?? "balanced"),
  };
}
