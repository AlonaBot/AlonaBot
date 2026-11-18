package manage

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type generationSummary struct {
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
	HasDB     bool      `json:"has_db"`
}

func (s *Service) generations(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	unlock := s.lock(id)
	defer unlock()
	in, err := s.installed(r.Context(), id)
	if err != nil {
		respond(w, nil, err)
		return
	}
	ids, err := s.opt.Store.Retained(r.Context(), id)
	if err != nil {
		respond(w, nil, err)
		return
	}
	active, err := s.opt.Store.Active(r.Context(), id)
	if err != nil {
		respond(w, nil, err)
		return
	}
	records, err := s.opt.Store.Journals(r.Context(), id)
	if err != nil {
		respond(w, nil, err)
		return
	}
	versions := make(map[string]string)
	for _, record := range records {
		if record.Phase == "committed" {
			versions[record.NewGeneration] = record.Version
		}
	}
	result := make([]generationSummary, 0, len(ids))
	for _, generation := range ids {
		created, err := s.opt.Store.GenerationCreatedAt(r.Context(), id, generation)
		if err != nil {
			respond(w, nil, err)
			return
		}
		version, exists := versions[generation]
		// 开发组件共用固定代次，程序版本随当前开发登记更新；安装包取提交时的程序快照。
		if in.Source == "dev" {
			version, exists = in.Version, true
		}
		if !exists {
			respond(w, nil, fmt.Errorf("generation %s has no committed program snapshot", generation))
			return
		}
		path := filepath.Join(filepath.Dir(s.opt.Packages), "plugins", id, "generations", generation, "db.sqlite")
		if in.Source == "dev" {
			path = filepath.Join(filepath.Dir(s.opt.Packages), "plugins", id, "dev.sqlite")
		}
		_, err = os.Stat(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			respond(w, nil, err)
			return
		}
		result = append(result, generationSummary{generation, version, created, generation == active, err == nil})
	}
	respond(w, struct {
		Generations []generationSummary `json:"generations"`
	}{result}, nil)
}
