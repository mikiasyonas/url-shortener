package gorm

import (
	"context"
	"log"
	"testing"

	"github.com/mikiasyonas/url-shortener/internal/core/domain"
	"github.com/mikiasyonas/url-shortener/pkg/config"
	"github.com/mikiasyonas/url-shortener/pkg/database"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

type URLRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo *URLRepository
	ctx  context.Context
}

func (suite *URLRepositoryTestSuite) SetupTest() {
	cfg := config.Load()

	if err := cfg.Validate(); err != nil {
		log.Fatal("❌ Invalid configuration:", err)
	}

	db, err := database.Connect(&cfg.Database)
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewURLRepository(db)
	suite.ctx = context.Background()

	suite.db.Exec("DELETE FROM urls")
}

func (suite *URLRepositoryTestSuite) TestSaveAndFind() {
	url, err := domain.NewURL("https://example.com", "abc123")
	suite.NoError(err)

	err = suite.repo.Save(suite.ctx, url)
	suite.NoError(err)

	found, err := suite.repo.FindByShortCode(suite.ctx, "abc123")
	suite.NoError(err)
	suite.Equal(url.OriginalURL, found.OriginalURL)
	suite.Equal(url.ShortCode, found.ShortCode)
}

func (suite *URLRepositoryTestSuite) TestSave_DuplicateShortCode() {
	url1, _ := domain.NewURL("https://example.com", "abc123")
	url2, _ := domain.NewURL("https://example.org", "abc123")

	err := suite.repo.Save(suite.ctx, url1)
	suite.NoError(err)

	err = suite.repo.Save(suite.ctx, url2)
	suite.ErrorIs(err, domain.ErrShortCodeTaken)
}

func (suite *URLRepositoryTestSuite) TestFindByShortCode_NotFound() {
	_, err := suite.repo.FindByShortCode(suite.ctx, "nonexistent")
	suite.ErrorIs(err, domain.ErrURLNotFound)
}

func (suite *URLRepositoryTestSuite) TestIncrementClickCount() {
	url, _ := domain.NewURL("https://example.com", "abc123")
	suite.repo.Save(suite.ctx, url)

	err := suite.repo.IncrementClickCount(suite.ctx, "abc123")
	suite.NoError(err)

	updated, err := suite.repo.FindByShortCode(suite.ctx, "abc123")
	suite.NoError(err)
	suite.Equal(int64(1), updated.ClickCount)
}

func TestURLRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(URLRepositoryTestSuite))
}
