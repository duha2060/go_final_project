package api

import (
	"errors"
	"net/http"
	"time"

	"go_final_project/internal/db"
)

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("метод не поддерживается"))
		return
	}

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

	if task.Repeat == "" {
		err = db.DeleteTask(id)
	} else {
		var next string
		next, err = NextDate(time.Now(), task.Date, task.Repeat)
		if err == nil {
			err = db.UpdateDate(next, id)
		}
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}
