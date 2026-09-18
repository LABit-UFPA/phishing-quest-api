-- Expande study_participants (criada minima na V20260917120000, so
-- user_id+consented_at) com os campos completos do desenho
-- experimental, e cria assessments para os instrumentos de
-- pre/pos/pos-tardio (ROADMAP_PESQUISA_2027.md, Fase 2 e 5).

ALTER TABLE phishing_quest.study_participants
    ADD COLUMN IF NOT EXISTS cohort_id UUID,
    ADD COLUMN IF NOT EXISTS condition VARCHAR(64),
    ADD COLUMN IF NOT EXISTS consent_version VARCHAR(32),
    ADD COLUMN IF NOT EXISTS demographics_json JSONB,
    ADD COLUMN IF NOT EXISTS withdrawn_at TIMESTAMPTZ;

COMMENT ON COLUMN phishing_quest.study_participants.cohort_id IS 'Identifica a turma/coorte do participante, para ranking e analise por grupo';
COMMENT ON COLUMN phishing_quest.study_participants.condition IS 'Condicao experimental atribuida no momento do consentimento (ex.: feedback_formativo, feedback_binario)';
COMMENT ON COLUMN phishing_quest.study_participants.consent_version IS 'Versao do TCLE aceita pelo participante — necessario para rastrear mudancas no termo ao longo do estudo';
COMMENT ON COLUMN phishing_quest.study_participants.demographics_json IS 'Dados demograficos autoinformados (idade, escolaridade, experiencia previa), coletados no momento do consentimento';
COMMENT ON COLUMN phishing_quest.study_participants.withdrawn_at IS 'Momento em que o participante retirou o consentimento (NULL enquanto ativo). Direito de exclusao (LGPD).';

CREATE TABLE IF NOT EXISTS phishing_quest.assessments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES phishing_quest.users(id),
    phase VARCHAR(16) NOT NULL,
    instrument_version VARCHAR(32) NOT NULL,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    responses_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_assessments_phase CHECK (phase IN ('pre', 'post', 'delayed_4w'))
);

COMMENT ON TABLE phishing_quest.assessments IS 'Respostas dos instrumentos de avaliacao (teste de deteccao, escalas de confianca/autoeficacia, SUS, IMI) coletadas nas fases pre-teste, pos-teste imediato e pos-tardio (4 semanas).';
COMMENT ON COLUMN phishing_quest.assessments.phase IS 'pre | post | delayed_4w';
COMMENT ON COLUMN phishing_quest.assessments.instrument_version IS 'Versao do instrumento aplicado, para rastrear mudancas nas formas paralelas A/B/C ao longo do estudo';
COMMENT ON COLUMN phishing_quest.assessments.responses_json IS 'Respostas brutas do instrumento (shape varia por tipo de instrumento e fase)';

CREATE INDEX IF NOT EXISTS idx_assessments_user_id ON phishing_quest.assessments(user_id);
CREATE INDEX IF NOT EXISTS idx_assessments_phase ON phishing_quest.assessments(phase);
