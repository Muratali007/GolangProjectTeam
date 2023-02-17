package main

import (
	"errors"
	"fmt"
	"net/http"
	"piscine/internal/data"
	"piscine/internal/validator"
)

func (app *application) createFootballerHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name            string   `json:"name"`
		Titles          int      `json:"titles"`
		StartedPlayYear int32    `json:"started_play_year"`
		Year            int32    `json:"year"`
		Club            string   `json:"club"`
		PlayedClubs     int      `json:"played_clubs"`
		Position        []string `json:"position"`
		Goals           int      `json:"goals"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	footballer := &data.Footballer{
		Name:            input.Name,
		Titles:          input.Titles,
		StartedPlayYear: input.StartedPlayYear,
		Year:            input.Year,
		Club:            input.Club,
		PlayedClubs:     input.PlayedClubs,
		Position:        input.Position,
		Goals:           input.Goals,
	}
	v := validator.New()

	if data.ValidateFootballer(v, footballer); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Footballers.Insert(footballer)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/footballer/%d", footballer.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"footballer": footballer}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showFootballerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	footballer, err := app.models.Footballers.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"footballer": footballer}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateFootballerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	footballer, err := app.models.Footballers.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name            *string   `json:"name"`
		Titles          *int      `json:"titles"`
		StartedPlayYear *int32    `json:"started_play_year"`
		Year            *int32    `json:"year"`
		Club            *string   `json:"club"`
		PlayedClubs     *int      `json:"played_clubs"`
		Position        []string `json:"position"`
		Goals           *int      `json:"goals"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.Name != nil {
		footballer.Name = *input.Name
	}
	if input.Titles != nil {
		footballer.Titles = *input.Titles
	}
	if input.StartedPlayYear != nil {
		footballer.StartedPlayYear = *input.StartedPlayYear
	}
	if input.Year != nil {
		footballer.Year = *input.Year
	}
	if input.Club != nil {
		footballer.Club = *input.Club
	}
	if input.PlayedClubs != nil {
		footballer.PlayedClubs = *input.PlayedClubs
	}
	if input.Position != nil {
		footballer.Position = input.Position
	}
	if input.Goals != nil {
		footballer.Goals = *input.Goals
	}

	v := validator.New()
	if data.ValidateFootballer(v, footballer); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Footballers.Update(footballer)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"footballer": footballer}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteFootballerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Footballers.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "footballer successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listFootballerHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name            string
		Club            string
		Position        []string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs,"names","")
	input.Club = app.readString(qs,"club","")

	input.Position = app.readCSV(qs,"positions",[]string{})

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")

	input.Filters.SortSafelist = []string{"id","names","titles","startedplayyear","year","goals","-id","-names","-titles","-startedplayyear","-year","-goals"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	footballers,metadata, err := app.models.Footballers.GetAll(input.Name,input.Club,input.Position,input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"footballers": footballers, "metadata":metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}