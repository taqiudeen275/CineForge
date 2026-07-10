import { expect, test } from "bun:test";
import {
  formValuesFromFormData,
  serializeProjectSettings,
  type ProjectSettings,
} from "./project-settings";

const baseSettings: ProjectSettings = {
  name: "Old cut",
  projectType: "single",
  productionFormat: "short_film",
  aspectWidth: 16,
  aspectHeight: 9,
  frameRateNumerator: 24,
  frameRateDenominator: 1,
  audioLanguage: "en",
  rating: "moderate",
  styleDirection: "",
  qualityPolicy: "balanced",
  costCeilingMicros: null,
};

test("serializes project settings form values into the API contract", () => {
  const serialized = serializeProjectSettings(baseSettings, {
    name: "Night Market",
    projectType: "series",
    productionFormat: "episodic",
    aspect: "21:9",
    fps: "24000/1001",
    language: "en-GH",
    rating: "mature",
    style: "Noir neon realism",
    quality: "final",
    costCeiling: "12.5",
  });

  expect(serialized).toEqual({
    ...baseSettings,
    name: "Night Market",
    projectType: "series",
    productionFormat: "episodic",
    aspectWidth: 21,
    aspectHeight: 9,
    frameRateNumerator: 24000,
    frameRateDenominator: 1001,
    audioLanguage: "en-GH",
    rating: "mature",
    styleDirection: "Noir neon realism",
    qualityPolicy: "final",
    costCeilingMicros: 12_500_000,
  });
});

test("project settings form defaults are safe when fields are absent", () => {
  const form = new FormData();
  const serialized = serializeProjectSettings(baseSettings, formValuesFromFormData(form));

  expect(serialized.aspectWidth).toBe(16);
  expect(serialized.aspectHeight).toBe(9);
  expect(serialized.frameRateNumerator).toBe(24);
  expect(serialized.frameRateDenominator).toBe(1);
  expect(serialized.audioLanguage).toBe("en");
  expect(serialized.rating).toBe("moderate");
  expect(serialized.qualityPolicy).toBe("balanced");
});
