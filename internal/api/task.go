package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"go_final_project/internal/db"
)

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("не указан идентификатор"))
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if task.ID == "" {
		writeError(w, http.StatusBadRequest, errors.New("не указан идентификатор"))
		return
	}
	if task.Title == "" {
		writeError(w, http.StatusBadRequest, errors.New("не указан заголовок задачи"))
		return
	}
	if err := checkDate(&task); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("не указан идентификатор"))
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}
