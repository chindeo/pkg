package control

import (
	"fmt"

	"github.com/chindeo/pkg/service"
)

func Start(srvName string) error {
	status, err := service.ServiceStatus(srvName)
	if err != nil {
		return fmt.Errorf("get service status  %w", err)
	}

	if status == service.StatusRunning {
		return nil
	}

	if status == service.StatusUninstall {
		return fmt.Errorf("service uninstall")
	}

	if status == service.StatusStopped {
		return service.ServiceStart(srvName)
	}

	return nil
}
