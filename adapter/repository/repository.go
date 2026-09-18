package repository

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type IRepository[T any] interface {
	Create(entity *T) (*T, error)
	Update(entity *T) (*T, error)
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (*T, error)
	GetAll() ([]*T, error)
}

type Repository[T any] struct {
	db *gorm.DB
}

func NewRepository[T any](db *gorm.DB) IRepository[T] {
	return &Repository[T]{db: db}
}

// describe monta a identificacao da entidade para log: tipo + id, e nada
// mais.
//
// Antes estas mensagens usavam "%+v", despejando a struct inteira. Como
// o domain.User carregava a senha em texto puro, todo cadastro gravava a
// senha do usuario no log em nivel INFO (issue #60). O campo de senha foi
// removido do dominio, e este helper impede que a proxima entidade com
// dado sensivel repita o problema: para vazar algo agora, alguem precisa
// escrever um log novo de proposito.
//
// A reflexao fica contida aqui (o repositorio e generico, nao ha
// interface comum com id) e nao entra em nenhum caminho de request:
// so alimenta log.
func describe(entity any) string {
	value := reflect.Indirect(reflect.ValueOf(entity))
	typeName := value.Type().Name()

	if value.Kind() == reflect.Struct {
		if field := value.FieldByName("Id"); field.IsValid() {
			if id, ok := field.Interface().(uuid.UUID); ok {
				return fmt.Sprintf("%s(id=%s)", typeName, id)
			}
		}
	}

	return typeName
}

func (r *Repository[T]) Create(entity *T) (*T, error) {
	logrus.Infof("Criando nova entidade: %s", describe(entity))
	if err := r.db.Create(entity).Error; err != nil {
		logrus.Errorf("Erro ao criar entidade: %v", err)
		return nil, err
	}
	logrus.Infof("Entidade criada com sucesso: %s", describe(entity))
	return entity, nil
}

func (r *Repository[T]) Update(entity *T) (*T, error) {
	logrus.Infof("Atualizando entidade: %s", describe(entity))
	if err := r.db.Save(entity).Error; err != nil {
		logrus.Errorf("Erro ao atualizar entidade: %v", err)
		return nil, err
	}
	logrus.Infof("Entidade atualizada com sucesso: %s", describe(entity))
	return entity, nil
}

func (r *Repository[T]) Delete(id uuid.UUID) error {
	logrus.Infof("Deletando entidade com ID: %s", id)
	var entity T
	if err := r.db.Delete(&entity, id).Error; err != nil {
		logrus.Errorf("Erro ao deletar entidade com ID %s: %v", id, err)
		return err
	}
	logrus.Infof("Entidade deletada com sucesso com ID: %s", id)
	return nil
}

func (r *Repository[T]) GetByID(id uuid.UUID) (*T, error) {
	logrus.Infof("Buscando entidade com ID: %s", id)
	var entity T
	if err := r.db.First(&entity, id).Error; err != nil {
		logrus.Errorf("Erro ao buscar entidade com ID %s: %v", id, err)
		return nil, err
	}
	logrus.Infof("Entidade encontrada com ID: %s", id)
	return &entity, nil
}

func (r *Repository[T]) GetAll() ([]*T, error) {
	logrus.Infof("Buscando todas as entidades")
	var entities []*T
	if err := r.db.Find(&entities).Error; err != nil {
		logrus.Errorf("Erro ao buscar todas as entidades: %v", err)
		return nil, err
	}
	logrus.Infof("Encontradas %d entidades", len(entities))
	return entities, nil
}
