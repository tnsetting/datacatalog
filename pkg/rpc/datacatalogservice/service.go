package datacatalogservice

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/lyft/datacatalog/pkg/manager/impl"
	"github.com/lyft/datacatalog/pkg/manager/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories"
	"github.com/lyft/datacatalog/pkg/repositories/config"
	"github.com/lyft/datacatalog/pkg/runtime"
	datacatalog "github.com/lyft/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
)

type DataCatalogService struct {
	DatasetManager  interfaces.DatasetManager
	ArtifactManager interfaces.ArtifactManager
	TagManager      interfaces.TagManager
}

// TODO: add metrics and counters to these service methods
func (s *DataCatalogService) CreateDataset(ctx context.Context, request *datacatalog.CreateDatasetRequest) (*datacatalog.CreateDatasetResponse, error) {
	return s.DatasetManager.CreateDataset(ctx, *request)
}

func (s *DataCatalogService) CreateArtifact(ctx context.Context, request *datacatalog.CreateArtifactRequest) (*datacatalog.CreateArtifactResponse, error) {
	return s.ArtifactManager.CreateArtifact(ctx, *request)
}

func (s *DataCatalogService) GetDataset(ctx context.Context, request *datacatalog.GetDatasetRequest) (*datacatalog.GetDatasetResponse, error) {
	return s.DatasetManager.GetDataset(ctx, *request)
}

func (s *DataCatalogService) GetArtifact(ctx context.Context, request *datacatalog.GetArtifactRequest) (*datacatalog.GetArtifactResponse, error) {
	return s.ArtifactManager.GetArtifact(ctx, *request)
}

func (s *DataCatalogService) AddTag(ctx context.Context, request *datacatalog.AddTagRequest) (*datacatalog.AddTagResponse, error) {
	return s.TagManager.AddTag(ctx, *request)
}

func NewDataCatalogService() *DataCatalogService {
	ctx := context.Background()

	dataCatalogScope := "datacatalog"
	catalogScope := promutils.NewScope(dataCatalogScope).NewSubScope("service")

	defer func() {
		if err := recover(); err != nil {
			catalogScope.MustNewCounter("initialization_panic",
				"panics encountered initializating the datacatalog service").Inc()
			logger.Fatalf(context.Background(), fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()

	storeConfig := storage.GetConfig()
	dataStorageClient, err := storage.NewDataStore(storeConfig, catalogScope.NewSubScope("storage"))
	if err != nil {
		logger.Errorf(ctx, "Failed to create DataStore %v, err %v", storeConfig, err)
		panic(err)
	}
	logger.Infof(ctx, "Created data storage.")

	configProvider := runtime.NewConfigurationProvider()
	baseStorageReference := dataStorageClient.GetBaseContainerFQN(ctx)
	dataCatalogConfig := configProvider.ApplicationConfiguration().GetDataCatalogConfig()
	storagePrefix, err := dataStorageClient.ConstructReference(ctx, baseStorageReference, dataCatalogConfig.StoragePrefix)
	if err != nil {
		logger.Errorf(ctx, "Failed to create prefix %v, err %v", dataCatalogConfig.StoragePrefix, err)
		panic(err)
	}

	dbConfigValues := configProvider.ApplicationConfiguration().GetDbConfig()
	dbConfig := config.DbConfig{
		Host:         dbConfigValues.Host,
		Port:         dbConfigValues.Port,
		DbName:       dbConfigValues.DbName,
		User:         dbConfigValues.User,
		Password:     dbConfigValues.Password,
		ExtraOptions: dbConfigValues.ExtraOptions,
	}
	repos := repositories.GetRepository(repositories.POSTGRES, dbConfig, catalogScope)
	logger.Infof(ctx, "Created DB connection.")

	return &DataCatalogService{
		DatasetManager:  impl.NewDatasetManager(repos, dataStorageClient, catalogScope.NewSubScope("dataset")),
		ArtifactManager: impl.NewArtifactManager(repos, dataStorageClient, storagePrefix, catalogScope.NewSubScope("artifact")),
		TagManager:      impl.NewTagManager(repos, dataStorageClient, catalogScope.NewSubScope("tag")),
	}
}