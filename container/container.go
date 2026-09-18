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

	ItemRepo             *repository.IItemRepository
	ItemUseCase          *usecase.ItemUseCase
	ItemHandler          *handler.ItemHandler
	ItemSelectionUseCase *usecase.ItemSelectionUseCase
	ItemSelectionHandler *handler.ItemSelectionHandler

	CueRepo     *repository.ICueRepository
	ItemCueRepo *repository.IItemCueRepository
	CueUseCase  *usecase.CueUseCase
	CueHandler  *handler.CueHandler

	AttemptRepo    *repository.IAttemptRepository
	ConsentRepo    *repository.IStudyParticipantRepository
	AttemptUseCase *usecase.AttemptUseCase
	AttemptHandler *handler.AttemptHandler

	TelemetryRepo    *repository.ITelemetryEventRepository
	TelemetryUseCase *usecase.TelemetryUseCase
	TelemetryHandler *handler.TelemetryHandler

	AssessmentRepo    *repository.IAssessmentRepository
	AssessmentUseCase *usecase.AssessmentUseCase
	AssessmentHandler *handler.AssessmentHandler

	ResearchExportRepo    *repository.IResearchExportRepository
	ResearchExportUseCase *usecase.ResearchExportUseCase
	ResearchExportHandler *handler.ResearchExportHandler

	UserStatsRepo    *repository.IUserStatsRepository
	UserStatsUseCase *usecase.UserStatsUseCase
	UserStatsHandler *handler.UserStatsHandler

	ReviewScheduleRepo    *repository.IReviewScheduleRepository
	ReviewScheduleUseCase *usecase.ReviewScheduleUseCase
	ReviewScheduleHandler *handler.ReviewScheduleHandler
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

	itemRepo := repository.NewItemRepository(db)
	itemUseCase := usecase.NewItemUseCase(itemRepo)
	itemHandler := handler.NewItemHandler(itemUseCase)

	// userStatsRepo precisa existir antes de itemSelectionUseCase, que
	// o usa para descobrir quais pistas o usuario ainda erra (selecao
	// adaptativa, issue #28).
	userStatsRepo := repository.NewUserStatsRepository(db)
	userStatsUseCase := usecase.NewUserStatsUseCase(userStatsRepo)
	userStatsHandler := handler.NewUserStatsHandler(userStatsUseCase)

	itemSelectionUseCase := usecase.NewItemSelectionUseCase(itemRepo, userStatsRepo)
	itemSelectionHandler := handler.NewItemSelectionHandler(itemSelectionUseCase)

	cueRepo := repository.NewCueRepository(db)
	itemCueRepo := repository.NewItemCueRepository(db)
	cueUseCase := usecase.NewCueUseCase(cueRepo, itemCueRepo, itemRepo)
	cueHandler := handler.NewCueHandler(cueUseCase)

	// reviewScheduleUseCase precisa existir antes de attemptUseCase,
	// que o usa para atualizar a fila de revisao espacada a cada
	// tentativa registrada.
	reviewScheduleRepo := repository.NewReviewScheduleRepository(db)
	reviewScheduleUseCase := usecase.NewReviewScheduleUseCase(reviewScheduleRepo, itemCueRepo)
	reviewScheduleHandler := handler.NewReviewScheduleHandler(reviewScheduleUseCase)

	attemptRepo := repository.NewAttemptRepository(db)
	consentRepo := repository.NewStudyParticipantRepository(db)
	attemptUseCase := usecase.NewAttemptUseCase(attemptRepo, itemRepo, consentRepo, reviewScheduleUseCase)
	attemptHandler := handler.NewAttemptHandler(attemptUseCase)

	telemetryRepo := repository.NewTelemetryEventRepository(db)
	telemetryUseCase := usecase.NewTelemetryUseCase(telemetryRepo)
	telemetryHandler := handler.NewTelemetryHandler(telemetryUseCase)

	assessmentRepo := repository.NewAssessmentRepository(db)
	assessmentUseCase := usecase.NewAssessmentUseCase(assessmentRepo)
	assessmentHandler := handler.NewAssessmentHandler(assessmentUseCase)

	researchExportRepo := repository.NewResearchExportRepository(db)
	researchExportUseCase := usecase.NewResearchExportUseCase(researchExportRepo)
	researchExportHandler := handler.NewResearchExportHandler(researchExportUseCase)

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

		ItemRepo:             &itemRepo,
		ItemUseCase:          itemUseCase,
		ItemHandler:          itemHandler,
		ItemSelectionUseCase: itemSelectionUseCase,
		ItemSelectionHandler: itemSelectionHandler,

		CueRepo:     &cueRepo,
		ItemCueRepo: &itemCueRepo,
		CueUseCase:  cueUseCase,
		CueHandler:  cueHandler,

		AttemptRepo:    &attemptRepo,
		ConsentRepo:    &consentRepo,
		AttemptUseCase: attemptUseCase,
		AttemptHandler: attemptHandler,

		TelemetryRepo:    &telemetryRepo,
		TelemetryUseCase: telemetryUseCase,
		TelemetryHandler: telemetryHandler,

		AssessmentRepo:    &assessmentRepo,
		AssessmentUseCase: assessmentUseCase,
		AssessmentHandler: assessmentHandler,

		ResearchExportRepo:    &researchExportRepo,
		ResearchExportUseCase: researchExportUseCase,
		ResearchExportHandler: researchExportHandler,

		UserStatsRepo:    &userStatsRepo,
		UserStatsUseCase: userStatsUseCase,
		UserStatsHandler: userStatsHandler,

		ReviewScheduleRepo:    &reviewScheduleRepo,
		ReviewScheduleUseCase: reviewScheduleUseCase,
		ReviewScheduleHandler: reviewScheduleHandler,
	}
}
