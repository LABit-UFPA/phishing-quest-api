package tests

import (
	"testing"

	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockResearchExportRepository implementa repository.IResearchExportRepository.
type MockResearchExportRepository struct {
	mock.Mock
}

func (m *MockResearchExportRepository) ListAttemptsWithItemInfo() ([]*domain.ResearchExportRow, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.ResearchExportRow), args.Error(1)
	}
	return nil, args.Error(1)
}

// TestResearchExportUseCase_Pseudonimiza garante que o UserId real
// NUNCA aparece no resultado exportado — e substituido por um hash.
func TestResearchExportUseCase_Pseudonimiza(t *testing.T) {
	t.Setenv("RESEARCH_EXPORT_SALT", "salt-de-teste")

	mockRepo := new(MockResearchExportRepository)
	uc := usecase.NewResearchExportUseCase(mockRepo)

	realUserID := uuid.New()
	mockRepo.On("ListAttemptsWithItemInfo").Return([]*domain.ResearchExportRow{
		{PseudoUserId: realUserID.String(), ItemId: uuid.New(), Channel: "email"},
	}, nil)

	rows, err := uc.ExportAttempts()

	assert.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.NotEqual(t, realUserID.String(), rows[0].PseudoUserId)
	assert.NotContains(t, rows[0].PseudoUserId, realUserID.String())
	assert.NotEmpty(t, rows[0].PseudoUserId)
}

// TestResearchExportUseCase_HashEstavel garante que o mesmo usuario
// (mesmo salt) sempre produz o mesmo pseudo-id — necessario para
// agrupar tentativas do mesmo participante ao longo do tempo na
// analise estatistica.
func TestResearchExportUseCase_HashEstavel(t *testing.T) {
	t.Setenv("RESEARCH_EXPORT_SALT", "salt-de-teste")

	mockRepo := new(MockResearchExportRepository)
	uc := usecase.NewResearchExportUseCase(mockRepo)

	realUserID := uuid.New().String()
	mockRepo.On("ListAttemptsWithItemInfo").Return([]*domain.ResearchExportRow{
		{PseudoUserId: realUserID, ItemId: uuid.New()},
		{PseudoUserId: realUserID, ItemId: uuid.New()},
	}, nil)

	rows, err := uc.ExportAttempts()

	assert.NoError(t, err)
	assert.Equal(t, rows[0].PseudoUserId, rows[1].PseudoUserId)
}

// TestResearchExportUseCase_HashDependeDoSalt garante que saltar com
// valores diferentes produz pseudo-ids diferentes para o mesmo
// usuario — sem o salt correto, quem so tem o CSV nao consegue
// recalcular/correlacionar a identidade real.
func TestResearchExportUseCase_HashDependeDoSalt(t *testing.T) {
	realUserID := uuid.New().String()

	// Duas instancias de usecase/mock independentes: cada uma devolve
	// uma copia fresca dos dados brutos (UserId real, ainda nao
	// pseudonimizado), simulando duas requisicoes HTTP separadas.
	mockRepoA := new(MockResearchExportRepository)
	mockRepoA.On("ListAttemptsWithItemInfo").Return([]*domain.ResearchExportRow{
		{PseudoUserId: realUserID, ItemId: uuid.New()},
	}, nil)
	ucA := usecase.NewResearchExportUseCase(mockRepoA)

	mockRepoB := new(MockResearchExportRepository)
	mockRepoB.On("ListAttemptsWithItemInfo").Return([]*domain.ResearchExportRow{
		{PseudoUserId: realUserID, ItemId: uuid.New()},
	}, nil)
	ucB := usecase.NewResearchExportUseCase(mockRepoB)

	t.Setenv("RESEARCH_EXPORT_SALT", "salt-a")
	rowsA, err := ucA.ExportAttempts()
	assert.NoError(t, err)

	t.Setenv("RESEARCH_EXPORT_SALT", "salt-b")
	rowsB, err := ucB.ExportAttempts()
	assert.NoError(t, err)

	assert.NotEqual(t, rowsA[0].PseudoUserId, rowsB[0].PseudoUserId)
}

func TestResearchExportUseCase_PropagaErroDoRepo(t *testing.T) {
	mockRepo := new(MockResearchExportRepository)
	uc := usecase.NewResearchExportUseCase(mockRepo)

	mockRepo.On("ListAttemptsWithItemInfo").Return(nil, assert.AnError)

	rows, err := uc.ExportAttempts()

	assert.Nil(t, rows)
	assert.Error(t, err)
}
