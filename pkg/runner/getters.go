package runner

import (
	"github.com/cohix/ragoo/pkg/embedder"
	"github.com/cohix/ragoo/pkg/service"
	"github.com/cohix/ragoo/pkg/storage"
)

func (r *Runner) embedder(ref string) embedder.Embedder {
	for _, emb := range r.config.Embedders {
		if emb.Name == ref {
			return embedder.EmbedderOfType(emb.Type, emb.Config)
		}
	}

	return nil
}

func (r *Runner) storage(ref string) storage.Storage {
	for _, str := range r.config.Storage {
		if str.Name == ref {
			return storage.StorageOfType(str.Type, str.Config)
		}
	}

	return nil
}

func (r *Runner) service(ref string) service.Service {
	for _, srv := range r.config.Services {
		if srv.Name == ref {
			return service.ServiceOfType(srv.Type, srv.Config)
		}
	}

	return nil
}
