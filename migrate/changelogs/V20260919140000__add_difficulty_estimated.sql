-- Separa a dificuldade ESTIMADA (a priori, pelo gerador da phishforge-api)
-- da dificuldade CALIBRADA (a posteriori, medida a partir da taxa de erro
-- real dos participantes -- issue #66 deste repositorio).
--
-- Ate aqui so existia difficulty_calibrated. Quando a integracao com a
-- phishforge-api (issue #62) mapear a resposta do gerador para items, a
-- estimativa (phishforge-api #9: derivada de phish_scale_cue_count +
-- phish_scale_premise_alignment no momento da geracao) acabaria gravada
-- na MESMA coluna que a calibracao usa -- a calibracao sobrescreveria o
-- proprio dado que deveria estar avaliando, e a comparacao estimado x
-- medido (resultado publicavel: "o gerador acerta o alvo de dificuldade
-- que lhe foi pedido?") seria perdida (issue #67).
--
-- Vocabulario fechado com CHECK em vez de convencao em COMMENT: os
-- COMMENTs desta tabela ja diziam "ex.: easy, medium, hard" e
-- "ex.: low, medium, high", mas nada impedia outro valor. Ingles,
-- alinhado com o que este repositorio ja usa para difficulty_calibrated
-- e phish_scale_premise_alignment -- a fronteira com o vocabulario
-- interno em portugues da phishforge-api (facil/medio/dificil,
-- baixo/medio/alto) e traduzida do lado de la (phishforge-api #2), nao
-- aqui.

ALTER TABLE phishing_quest.items
    ADD COLUMN difficulty_estimated VARCHAR(16);

-- phish_scale_premise_alignment estava em VARCHAR(32), maior do que
-- qualquer valor do vocabulario fechado (low/medium/high, no maximo 6
-- caracteres) -- dimensionado antes de o enum existir. Alinhado ao
-- mesmo tamanho de difficulty_calibrated/difficulty_estimated.
ALTER TABLE phishing_quest.items
    ALTER COLUMN phish_scale_premise_alignment TYPE VARCHAR(16);

ALTER TABLE phishing_quest.items
    ADD CONSTRAINT chk_items_difficulty_estimated
        CHECK (difficulty_estimated IS NULL OR difficulty_estimated IN ('easy', 'medium', 'hard')),
    ADD CONSTRAINT chk_items_difficulty_calibrated
        CHECK (difficulty_calibrated IS NULL OR difficulty_calibrated IN ('easy', 'medium', 'hard')),
    ADD CONSTRAINT chk_items_premise_alignment
        CHECK (phish_scale_premise_alignment IS NULL OR phish_scale_premise_alignment IN ('low', 'medium', 'high'));

COMMENT ON COLUMN phishing_quest.items.difficulty_estimated IS 'Dificuldade ESTIMADA a priori pelo gerador (phishforge-api #9), derivada de phish_scale_cue_count + phish_scale_premise_alignment no momento da criacao do item. Distinta de difficulty_calibrated (medida a posteriori a partir da taxa de erro real dos participantes -- ver issue #66). NUNCA gravar a estimativa em difficulty_calibrated. Vocabulario: easy | medium | hard.';
COMMENT ON COLUMN phishing_quest.items.difficulty_calibrated IS 'Dificuldade CALIBRADA a posteriori, medida a partir da taxa de erro real dos participantes (issue #66). Distinta de difficulty_estimated (a priori, do gerador). Vocabulario: easy | medium | hard.';
COMMENT ON COLUMN phishing_quest.items.phish_scale_premise_alignment IS 'Alinhamento da premissa do NIST Phish Scale com o contexto do usuario, julgado no momento da geracao (phishforge-api #9). Vocabulario: low | medium | high.';
