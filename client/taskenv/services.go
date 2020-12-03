package taskenv

import (
	"github.com/hashicorp/nomad/nomad/structs"
)

// InterpolateServices returns an interpolated copy of services and checks with
// values from the task's environment.
func InterpolateServices(taskEnv *TaskEnv, services []*structs.Service) []*structs.Service {
	// Guard against not having a valid taskEnv. This can be the case if the
	// PreKilling or Exited hook is run before Poststart.
	if taskEnv == nil || len(services) == 0 {
		return nil
	}

	interpolated := make([]*structs.Service, len(services))

	for i, origService := range services {
		// Create a copy as we need to reinterpolate every time the
		// environment changes
		service := origService.Copy()

		for _, check := range service.Checks {
			check.Name = taskEnv.ReplaceEnv(check.Name)
			check.Type = taskEnv.ReplaceEnv(check.Type)
			check.Command = taskEnv.ReplaceEnv(check.Command)
			check.Args = taskEnv.ParseAndReplace(check.Args)
			check.Path = taskEnv.ReplaceEnv(check.Path)
			check.Protocol = taskEnv.ReplaceEnv(check.Protocol)
			check.PortLabel = taskEnv.ReplaceEnv(check.PortLabel)
			check.InitialStatus = taskEnv.ReplaceEnv(check.InitialStatus)
			check.Method = taskEnv.ReplaceEnv(check.Method)
			check.GRPCService = taskEnv.ReplaceEnv(check.GRPCService)
			if len(check.Header) > 0 {
				header := make(map[string][]string, len(check.Header))
				for k, vs := range check.Header {
					newVals := make([]string, len(vs))
					for i, v := range vs {
						newVals[i] = taskEnv.ReplaceEnv(v)
					}
					header[taskEnv.ReplaceEnv(k)] = newVals
				}
				check.Header = header
			}
		}

		service.Name = taskEnv.ReplaceEnv(service.Name)
		service.PortLabel = taskEnv.ReplaceEnv(service.PortLabel)
		service.Tags = taskEnv.ParseAndReplace(service.Tags)
		service.CanaryTags = taskEnv.ParseAndReplace(service.CanaryTags)

		// interpolate connect
		service.Connect = interpolateConnect(taskEnv, service.Connect)

		if len(service.Meta) > 0 {
			meta := make(map[string]string, len(service.Meta))
			for k, v := range service.Meta {
				meta[k] = taskEnv.ReplaceEnv(v)
			}
			service.Meta = meta
		}

		if len(service.CanaryMeta) > 0 {
			canaryMeta := make(map[string]string, len(service.CanaryMeta))
			for k, v := range service.CanaryMeta {
				canaryMeta[k] = taskEnv.ReplaceEnv(v)
			}
			service.CanaryMeta = canaryMeta
		}

		interpolated[i] = service
	}

	return interpolated
}

func interpolateConnect(taskEnv *TaskEnv, orig *structs.ConsulConnect) *structs.ConsulConnect {
	if orig == nil {
		return nil
	}

	// make one copy and interpolate the rest in-place
	COPY := orig.Copy()

	if COPY.Gateway != nil {
		interpolateConnectGatewayProxy(taskEnv, COPY.Gateway.Proxy)
		interpolateConnectGatewayIngress(taskEnv, COPY.Gateway.Ingress)
	}

	if COPY.SidecarService != nil {
		//
	}

	if COPY.SidecarService != nil {
		//
	}

	return COPY
}

func interpolateConnectGatewayProxy(taskEnv *TaskEnv, proxy *structs.ConsulGatewayProxy) {
	if proxy == nil {
		return
	}

	for _, address := range proxy.EnvoyGatewayBindAddresses {
		address.Address = taskEnv.ReplaceEnv(address.Address)
	}
}

func interpolateConnectGatewayIngress(taskEnv *TaskEnv, ingress *structs.ConsulIngressConfigEntry) {
	if ingress == nil {
		return
	}

	for _, listener := range ingress.Listeners {
		listener.Protocol = taskEnv.ReplaceEnv(listener.Protocol)
		for _, service := range listener.Services {
			service.Name = taskEnv.ReplaceEnv(service.Name)
			service.Hosts = taskEnv.ParseAndReplace(service.Hosts)
		}
	}
}
