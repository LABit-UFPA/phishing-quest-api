-- Fila de revisao espacada (algoritmo Leitner por pista).
-- ROADMAP_PESQUISA_2027.md, Fase 3: pratica distribuida e repetida e
-- a resposta direta aos achados negativos da literatura sobre
-- treinamento de dose unica.

CREATE TABLE IF NOT EXISTS phishing_quest.review_schedule (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES phishing_quest.users(id),
    item_id UUID NOT NULL REFERENCES phishing_quest.items(id),
    cue_id UUID NOT NULL REFERENCES phishing_quest.cues(id),
    box INT NOT NULL DEFAULT 1,
    due_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_result BOOLEAN,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_review_schedule_user_item_cue UNIQUE (user_id, item_id, cue_id),
    CONSTRAINT chk_review_schedule_box CHECK (box BETWEEN 1 AND 5)
);

COMMENT ON TABLE phishing_quest.review_schedule IS 'Agendamento de revisao espacada (Leitner) por combinacao usuario+item+pista. Item errado reaparece antes (volta pra box 1); item dominado espaca no tempo (avanca de box).';
COMMENT ON COLUMN phishing_quest.review_schedule.box IS 'Caixa do Leitner (1 a 5). Box maior = intervalo de repeticao maior.';
COMMENT ON COLUMN phishing_quest.review_schedule.due_at IS 'Momento a partir do qual este item+pista volta a ser elegivel para revisao';
COMMENT ON COLUMN phishing_quest.review_schedule.last_result IS 'Se a ultima tentativa envolvendo este item+pista foi correta';

CREATE INDEX IF NOT EXISTS idx_review_schedule_user_due ON phishing_quest.review_schedule(user_id, due_at);
