package httpapi

import (
	"testing"
	"github.com/ats-tech/cineforge/services/control-api/internal/domain"
)

func TestValidateProject(t *testing.T){
	valid:=domain.ProjectSettings{Name:"Story",ProjectType:"single",ProductionFormat:"short_film",AspectWidth:16,AspectHeight:9,FrameRateNumerator:24,FrameRateDenominator:1,AudioLanguage:"en",Rating:"moderate",QualityPolicy:"balanced"}
	if !validateProject(valid){t.Fatal("expected valid project")}
	valid.ProjectType="invalid";if validateProject(valid){t.Fatal("invalid project type must fail")}
}
