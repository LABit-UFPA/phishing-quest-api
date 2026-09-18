package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"
)

type ResearchExportUseCase struct {
	exportRepo repository.IResearchExportRepository
}

func NewResearchExportUseCase(exportRepo repository.IResearchExportRepository) *ResearchExportUseCase {
	return &ResearchExportUseCase{exportRepo: exportRepo}
}

// ExportAttempts retorna as tentativas com o UserId real substituido
// por um hash estavel (SHA-256 truncado, com salt via
// RESEARCH_EXPORT_SALT). O hash e determinstico: o mesmo usuario
// sempre produz o mesmo PseudoUserId, o que permite agrupar tentativas
// do mesmo participante ao longo do tempo para analise longitudinal
// SEM expor a identidade real no arquivo exportado. Sem o salt (nao
// versionado, guardado so em variavel de ambiente), o hash nao pode
// ser recalculado/correlacionado por quem nao tem acesso ao banco.
func (reuc *ResearchExportUseCase) ExportAttempts() ([]*domain.ResearchExportRow, error) {
	rows, err := reuc.exportRepo.ListAttemptsWithItemInfo()
	if err != nil {
		return nil, err
	}

	salt := os.Getenv("RESEARCH_EXPORT_SALT")
	// Cria uma NOVA struct por linha em vez de mutar a recebida do
	// repositorio: mutar in-place faria uma segunda chamada (mesmo
	// slice/ponteiros, como em testes com mocks reutilizados) hashear
	// um valor ja pseudonimizado anteriormente, em vez do UserId real.
	pseudonymized := make([]*domain.ResearchExportRow, len(rows))
	for i, row := range rows {
		copyRow := *row
		copyRow.PseudoUserId = pseudonymize(row.PseudoUserId, salt)
		pseudonymized[i] = &copyRow
	}
	return pseudonymized, nil
}

func pseudonymize(realUserID, salt string) string {
	hash := sha256.Sum256([]byte(salt + realUserID))
	return hex.EncodeToString(hash[:])[:16]
}
