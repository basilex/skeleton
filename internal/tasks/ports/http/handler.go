package http

import (
	"encoding/json"
	"net/http"

	"github.com/basilex/skeleton/internal/tasks/application/command"
	"github.com/basilex/skeleton/internal/tasks/application/query"
	"github.com/basilex/skeleton/internal/tasks/domain"
)

type Handler struct {
	createTaskHandler         *command.CreateTaskHandler
	cancelTaskHandler         *command.CancelTaskHandler
	retryDeadLetterHandler    *command.RetryDeadLetterHandler
	createScheduleHandler     *command.CreateScheduleHandler
	deleteScheduleHandler     *command.DeleteScheduleHandler
	activateScheduleHandler   *command.ActivateScheduleHandler
	deactivateScheduleHandler *command.DeactivateScheduleHandler
	getTaskHandler            *query.GetTaskHandler
	listTasksHandler          *query.ListTasksHandler
	getActiveTasksHandler     *query.GetActiveTasksHandler
	getTaskStatsHandler       *query.GetTaskStatsHandler
	getScheduleHandler        *query.GetScheduleHandler
	listSchedulesHandler      *query.ListSchedulesHandler
	listDeadLettersHandler    *query.ListDeadLettersHandler
}

func NewHandler(
	createTaskHandler *command.CreateTaskHandler,
	cancelTaskHandler *command.CancelTaskHandler,
	retryDeadLetterHandler *command.RetryDeadLetterHandler,
	createScheduleHandler *command.CreateScheduleHandler,
	deleteScheduleHandler *command.DeleteScheduleHandler,
	activateScheduleHandler *command.ActivateScheduleHandler,
	deactivateScheduleHandler *command.DeactivateScheduleHandler,
	getTaskHandler *query.GetTaskHandler,
	listTasksHandler *query.ListTasksHandler,
	getActiveTasksHandler *query.GetActiveTasksHandler,
	getTaskStatsHandler *query.GetTaskStatsHandler,
	getScheduleHandler *query.GetScheduleHandler,
	listSchedulesHandler *query.ListSchedulesHandler,
	listDeadLettersHandler *query.ListDeadLettersHandler,
) *Handler {
	return &Handler{
		createTaskHandler:         createTaskHandler,
		cancelTaskHandler:         cancelTaskHandler,
		retryDeadLetterHandler:    retryDeadLetterHandler,
		createScheduleHandler:     createScheduleHandler,
		deleteScheduleHandler:     deleteScheduleHandler,
		activateScheduleHandler:   activateScheduleHandler,
		deactivateScheduleHandler: deactivateScheduleHandler,
		getTaskHandler:            getTaskHandler,
		listTasksHandler:          listTasksHandler,
		getActiveTasksHandler:     getActiveTasksHandler,
		getTaskStatsHandler:       getTaskStatsHandler,
		getScheduleHandler:        getScheduleHandler,
		listSchedulesHandler:      listSchedulesHandler,
		listDeadLettersHandler:    listDeadLettersHandler,
	}
}

// CreateTask handles POST /api/v1/tasks
// @Summary Create a new task
// @Description Create a new background task
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body CreateTaskRequest true "Task creation request"
// @Success 201 {object} CreateTaskResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tasks [post]
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	taskType, err := domain.ParseTaskType(req.TaskType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	priority := domain.TaskPriorityNormal
	if req.Priority != "" {
		priority, err = domain.ParseTaskPriority(req.Priority)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	cmd := command.CreateTaskCommand{
		TaskType:    taskType,
		Payload:     req.Payload,
		Priority:    priority,
		ScheduledAt: req.ScheduledAt,
		MaxAttempts: req.MaxAttempts,
	}

	taskID, err := h.createTaskHandler.Handle(r.Context(), cmd)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, CreateTaskResponse{TaskID: taskID.String()})
}

// GetTask handles GET /api/v1/tasks/:id
// @Summary Get task details
// @Description Get details of a specific task
// @Tags tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} TaskResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tasks/{id} [get]
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	if taskID == "" {
		respondWithError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	task, err := h.getTaskHandler.Handle(r.Context(), query.GetTaskQuery{
		ID: domain.TaskID(taskID),
	})
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, taskToResponse(task))
}

// ListTasks handles GET /api/v1/tasks
// @Summary List tasks
// @Description List tasks with optional filters
// @Tags tasks
// @Produce json
// @Param type query string false "Task type filter"
// @Param status query string false "Task status filter"
// @Param priority query string false "Task priority filter"
// @Param limit query int false "Limit"
// @Success 200 {array} TaskResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tasks [get]
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	var taskType *domain.TaskType
	if t := r.URL.Query().Get("type"); t != "" {
		parsed, err := domain.ParseTaskType(t)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		taskType = &parsed
	}

	var status *domain.TaskStatus
	if s := r.URL.Query().Get("status"); s != "" {
		parsed, err := domain.ParseTaskStatus(s)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		status = &parsed
	}

	var priority *domain.TaskPriority
	if p := r.URL.Query().Get("priority"); p != "" {
		parsed, err := domain.ParseTaskPriority(p)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		priority = &parsed
	}

	limit := 100

	tasks, err := h.listTasksHandler.Handle(r.Context(), query.ListTasksQuery{
		TaskType: taskType,
		Status:   status,
		Priority: priority,
		Limit:    limit,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = taskToResponse(task)
	}

	respondWithJSON(w, http.StatusOK, responses)
}

// CreateSchedule handles POST /api/v1/tasks/schedules
// @Summary Create a schedule
// @Description Create a scheduled task
// @Tags schedules
// @Accept json
// @Produce json
// @Param request body CreateScheduleRequest true "Schedule creation request"
// @Success 201 {object} CreateScheduleResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tasks/schedules [post]
func (h *Handler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	taskType, err := domain.ParseTaskType(req.TaskType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	cmd := command.CreateScheduleCommand{
		Name:     req.Name,
		TaskType: taskType,
		Payload:  req.Payload,
		Cron:     req.Cron,
		Timezone: req.Timezone,
	}

	scheduleID, err := h.createScheduleHandler.Handle(r.Context(), cmd)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, CreateScheduleResponse{ScheduleID: scheduleID.String()})
}

// ListSchedules handles GET /api/v1/tasks/schedules
// @Summary List schedules
// @Description List all task schedules
// @Tags schedules
// @Produce json
// @Success 200 {array} ScheduleResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tasks/schedules [get]
func (h *Handler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.listSchedulesHandler.Handle(r.Context(), query.ListSchedulesQuery{})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]ScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		responses[i] = scheduleToResponse(schedule)
	}

	respondWithJSON(w, http.StatusOK, responses)
}

// ListDeadLetters handles GET /api/v1/tasks/dead-letters
// @Summary List dead letters
// @Description List failed tasks in dead letter queue
// @Tags dead-letters
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} DeadLetterResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tasks/dead-letters [get]
func (h *Handler) ListDeadLetters(w http.ResponseWriter, r *http.Request) {
	limit := 100
	offset := 0

	deadLetters, err := h.listDeadLettersHandler.Handle(r.Context(), query.ListDeadLettersQuery{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]DeadLetterResponse, len(deadLetters))
	for i, dl := range deadLetters {
		responses[i] = deadLetterToResponse(dl)
	}

	respondWithJSON(w, http.StatusOK, responses)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}
