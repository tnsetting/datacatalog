package repositories

import (
	"github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/gormimpl"
	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/jinzhu/gorm"
)

type PostgresRepo struct {
	datasetRepo  interfaces.DatasetRepo
	artifactRepo interfaces.ArtifactRepo
	tagRepo      interfaces.TagRepo
}

func (dc *PostgresRepo) DatasetRepo() interfaces.DatasetRepo {
	return dc.datasetRepo
}

func (dc *PostgresRepo) ArtifactRepo() interfaces.ArtifactRepo {
	return dc.artifactRepo
}

func (dc *PostgresRepo) TagRepo() interfaces.TagRepo {
	return dc.tagRepo
}

func NewPostgresRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.DataCatalogRepo {
	return &PostgresRepo{
		datasetRepo:  gormimpl.NewDatasetRepo(db, errorTransformer, scope.NewSubScope("dataset")),
		artifactRepo: gormimpl.NewArtifactRepo(db, errorTransformer, scope.NewSubScope("artifact")),
		tagRepo:      gormimpl.NewTagRepo(db, errorTransformer, scope.NewSubScope("tag")),
	}
}
