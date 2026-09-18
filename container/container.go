package container

import (
	"phishing-quest/adapter/http/handler"
	"phishing-quest/adapter/repository"
	"phishing-quest/core/service"
	"phishing-quest/core/usecase"
	"phishing-quest/postgres"

	"gorm.io/gorm"
)

type Container struct {
	DB          *gorm.DB
	JWTService  service.IJWTService
	UserRepo    *repository.IUserRepository
	UserUseCase *usecase.UserUseCase
	UserHandler *handler.UserHandler

	CategoryRepo    *repository.ICategoryRepository
	CategoryUseCase *usecase.CategoryUseCase
	CategoryHandler *handler.CategoryHandler

	QuestionRepo    *repository.IQuestionRepository
	QuestionUseCase *usecase.QuestionUseCase
	QuestionHandler *handler.QuestionHandler

	AnswerRepo    *repository.IAnswerRepository
	AnswerUseCase *usecase.AnswerUseCase
	AnswerHandler *handler.AnswerHandler

	UserScoreRepo *repository.IUserScoreRepository

	UserAnswerRepo    *repository.IUserAnswerRepository
	UserAnswerUseCase *usecase.UserAnswerUseCase
	UserAnswerHandler *handler.UserAnswerHandler

	GameUseCase *usecase.GameUseCase
	GameHandler *handler.GameHandler

	RankingRepo    *repository.IRankingRepository
	RankingUseCase *usecase.RankingUseCase
	RankingHandler *handler.RankingHandler
}

func NewContainer() *Container {
	db := postgres.InitDB()
	jwtService := service.NewJWTService()

	userRepo := repository.NewUserRepository(db)
	userUseCase := usecase.NewUserUseCase(userRepo, jwtService)
	userHandler := handler.NewUserHandler(userUseCase)

	categoryRepo := repository.NewCategoryRepository(db)

	answerRepo := repository.NewAnswerRepository(db)
	answerUseCase := usecase.NewAnswerUseCase(answerRepo)
	answerHandler := handler.NewAnswerHandler(answerUseCase)

	// questionRepo precisa existir antes de CategoryUseCase, que depende
	// dele para listar questoes por categoria (GetQuestionsByCategoryID).
	questionRepo := repository.NewQuestionRepository(db)
	questionUseCase := usecase.NewQuestionUseCase(questionRepo, answerRepo)
	questionHandler := handler.NewQuestionHandler(questionUseCase)

	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo, questionRepo)
	CategoryHandler := handler.NewCategoryHandler(categoryUseCase)

	userScoreRepo := repository.NewUserScoreRepository(db)

	userAnswerRepo := repository.NewUserAnswerRepository(db)
	userAnswerUseCase := usecase.NewUserAnswerUseCase(userAnswerRepo)
	userAnswerHandler := handler.NewUserAnswerHandler(userAnswerUseCase)

	gameUseCase := usecase.NewGameUseCase(answerRepo, userRepo, userScoreRepo, userAnswerRepo)
	gameHandler := handler.NewGameHandler(gameUseCase)

	rankingRepo := repository.NewRankingRepository(db)
	rankingUseCase := usecase.NewRankingUseCase(rankingRepo)
	rankingHandler := handler.NewRankingHandler(rankingUseCase)

	return &Container{
		DB:          db,
		JWTService:  jwtService,
		UserRepo:    &userRepo,
		UserUseCase: userUseCase,
		UserHandler: userHandler,

		CategoryRepo:    &categoryRepo,
		CategoryUseCase: categoryUseCase,
		CategoryHandler: CategoryHandler,

		QuestionRepo:    &questionRepo,
		QuestionUseCase: questionUseCase,
		QuestionHandler: questionHandler,

		AnswerRepo:    &answerRepo,
		AnswerUseCase: answerUseCase,
		AnswerHandler: answerHandler,

		UserScoreRepo: &userScoreRepo,

		UserAnswerRepo:    &userAnswerRepo,
		UserAnswerUseCase: userAnswerUseCase,
		UserAnswerHandler: userAnswerHandler,

		GameUseCase: gameUseCase,
		GameHandler: gameHandler,

		RankingRepo:    &rankingRepo,
		RankingUseCase: rankingUseCase,
		RankingHandler: rankingHandler,
	}
}
