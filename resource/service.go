package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Service struct {
	Title     string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta      meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Service   string  `json:"-" yaml:"-"`
	Enabled   matcher `json:"enabled" yaml:"enabled"`
	Running   matcher `json:"running" yaml:"running"`
	Skip      bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
	RunLevels matcher `json:"runlevels,omitempty" yaml:"runlevels,omitempty"`
}

func (s *Service) ID() string      { return s.Service }
func (s *Service) SetID(id string) { s.Service = id }

func (s *Service) GetTitle() string { return s.Title }
func (s *Service) GetMeta() meta    { return s.Meta }

func (s *Service) Validate(sys *system.System) []TestResult {
	skip := false
	sysservice := sys.NewService(s.Service, sys, util.Config{})

	if s.Skip {
		skip = true
	}

	var results []TestResult
	if s.Enabled != nil {
		results = append(results, ValidateValue(s, "enabled", s.Enabled, sysservice.Enabled, skip))
	}
	if s.Running != nil {
		results = append(results, ValidateValue(s, "running", s.Running, sysservice.Running, skip))
	}
	if s.RunLevels != nil {
		results = append(results, ValidateValue(s, "runlevels", s.RunLevels, sysservice.RunLevels, skip))
	}
	return results
}

func NewService(sysService system.Service, config util.Config) (*Service, error) {
	service := sysService.Service()
	enabled, err := sysService.Enabled()
	if err != nil {
		return nil, err
	}
	running, err := sysService.Running()
	if err != nil {
		return nil, err
	}
	return &Service{
		Service: service,
		Enabled: enabled,
		Running: running,
	}, nil
}
