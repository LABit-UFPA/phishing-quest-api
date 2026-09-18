-- Taxonomia de pistas de phishing (cues) e sua associacao com items.
-- Base para o diagnostico por pista e a selecao adaptativa
-- (ROADMAP_PESQUISA_2027.md, Fase 2/3).

CREATE TABLE IF NOT EXISTS phishing_quest.cues (
    id UUID PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    label_pt VARCHAR(255) NOT NULL,
    category VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE phishing_quest.cues IS 'Taxonomia de pistas de phishing usadas para anotar items e diagnosticar o desempenho do usuario por tipo de pista';
COMMENT ON COLUMN phishing_quest.cues.code IS 'Identificador estavel usado no codigo (ex.: typosquat, urgency)';
COMMENT ON COLUMN phishing_quest.cues.category IS 'Agrupamento amplo da pista (ex.: technical, psychological)';

CREATE TABLE IF NOT EXISTS phishing_quest.item_cues (
    id UUID PRIMARY KEY,
    item_id UUID NOT NULL REFERENCES phishing_quest.items(id),
    cue_id UUID NOT NULL REFERENCES phishing_quest.cues(id),
    span_start INT,
    span_end INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_item_cues_item_cue UNIQUE (item_id, cue_id)
);

COMMENT ON TABLE phishing_quest.item_cues IS 'Associa um item (email/sms/whatsapp/etc) a uma ou mais pistas presentes nele, com o trecho (span) onde a pista aparece no conteudo';
COMMENT ON COLUMN phishing_quest.item_cues.span_start IS 'Posicao inicial (offset de caractere) da pista no conteudo do item, se aplicavel';
COMMENT ON COLUMN phishing_quest.item_cues.span_end IS 'Posicao final (offset de caractere) da pista no conteudo do item, se aplicavel';

CREATE INDEX IF NOT EXISTS idx_item_cues_item_id ON phishing_quest.item_cues(item_id);
CREATE INDEX IF NOT EXISTS idx_item_cues_cue_id ON phishing_quest.item_cues(cue_id);

-- Seed inicial das pistas (taxonomia definida no roadmap de pesquisa).
INSERT INTO phishing_quest.cues (id, code, label_pt, category) VALUES
    ('00000000-0000-0000-0000-000000000001', 'sender_domain_mismatch', 'Domínio do remetente não corresponde à organização', 'technical'),
    ('00000000-0000-0000-0000-000000000002', 'typosquat', 'Domínio com erro de digitação proposital (typosquatting)', 'technical'),
    ('00000000-0000-0000-0000-000000000003', 'homoglyph', 'Caractere visualmente semelhante usado para enganar (homóglifo)', 'technical'),
    ('00000000-0000-0000-0000-000000000004', 'urgency', 'Apelo à urgência ou prazo curto', 'psychological'),
    ('00000000-0000-0000-0000-000000000005', 'authority', 'Apelo à autoridade (banco, governo, chefia)', 'psychological'),
    ('00000000-0000-0000-0000-000000000006', 'generic_greeting', 'Saudação genérica, sem personalização', 'psychological'),
    ('00000000-0000-0000-0000-000000000007', 'credential_request', 'Solicitação direta de senha ou dado sensível', 'technical'),
    ('00000000-0000-0000-0000-000000000008', 'link_text_mismatch', 'Texto do link não corresponde ao destino real', 'technical'),
    ('00000000-0000-0000-0000-000000000009', 'unexpected_attachment', 'Anexo inesperado ou fora do contexto', 'technical'),
    ('00000000-0000-0000-0000-000000000010', 'scarcity', 'Apelo à escassez ou oferta por tempo limitado', 'psychological')
ON CONFLICT (id) DO NOTHING;
