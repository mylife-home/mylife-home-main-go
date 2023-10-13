package engine

import (
	"mylife-home-core-plugins-driver-absoluta/engine/itv2"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/commands"
	"reflect"
	"time"
)

type Service struct {
	client           *itv2.Client
	state            *State
	connectedChanged func(bool)
}

func NewService(serverAddress string, pin string, state *State, connectedChanged func(bool)) *Service {

	svc := &Service{
		client:           itv2.MakeClient(serverAddress, pin),
		state:            state,
		connectedChanged: connectedChanged,
	}

	svc.client.RegisterNotifications(svc.handleNotification)

	svc.client.RegisterStatusChange(func(status itv2.ConnectionStatus) {
		logger.Debugf("Connection status changed to %s", status)

		switch status {
		case itv2.ConnectionOpen:
			svc.connectedChanged(true)
		case itv2.ConnectionClosed:
			svc.connectedChanged(false)
		}
	})

	svc.client.RegisterStatusChange(func(status itv2.ConnectionStatus) {
		if status == itv2.ConnectionOpen {
			// We are inside a goroutine already
			svc.connectionLoop()
		}
	})

	return svc
}

func (svc *Service) Terminate() {
	svc.client.Close()
}

func (svc *Service) connectionLoop() {
	client := svc.client
	state := svc.state

	for {
		time.Sleep(time.Second * 3)

		if client.Status() != itv2.ConnectionOpen {
			return
		}

		if statuses, err := client.GetPartitionStatus(); err != nil {
			if client.Status() != itv2.ConnectionOpen {
				return
			}

			logger.Errorf("Error reading partition statuses: %s", err)
		} else {
			state.updatePartitionStatuses(statuses.GetData())
		}

		if statuses, err := client.GetZoneStatuses(); err != nil {
			if client.Status() != itv2.ConnectionOpen {
				return
			}

			logger.Errorf("Error reading zone statuses: %s", err)
		} else {
			state.updateZoneStatuses(statuses.GetData())
		}
	}
}

func (svc *Service) handleNotification(cmd commands.Command) {
	client := svc.client
	state := svc.state

	switch cmd := cmd.(type) {
	case *commands.SystemCapabilities:
		partitionCount := int(cmd.MaxPartitions.GetUint())
		zoneCount := int(cmd.MaxZones.GetUint())
		logger.Debugf("Got SystemCapabilities: partitions count=%d, zones count=%d", partitionCount, zoneCount)
		state.updateCapabilities(partitionCount, zoneCount)

	case *commands.PartitionAssignmentConfiguration:
		assignedPartitions := cmd.GetAssignedPartitions()
		logger.Debugf("Got PartitionAssignmentConfiguration: %+v", assignedPartitions)
		state.updateAssignedPartitions(assignedPartitions)

		for _, index := range assignedPartitions {
			go func(index int) {
				label, err := client.GetPartitionLabel(index)
				if err != nil {
					logger.Errorf("Error reading partition label %d: %s\n", index, err)
					return
				}

				state.updatePartitionLabel(index, label)
			}(index)
		}

	case *commands.ZoneAssignmentConfiguration:
		assignedZones := cmd.GetAssignedZones()
		// Note: cmd.Req.PartitionNumber.GetUint() seems unused (always 0)
		logger.Debugf("Got ZoneAssignmentConfiguration: %+v", assignedZones)
		state.updateAssignedZones(assignedZones)

		for _, index := range assignedZones {
			go func(index int) {
				label, err := client.GetZoneLabel(index)
				if err != nil {
					logger.Errorf("Error reading zone label %d: %s\n", index, err)
					return
				}

				state.updateZoneLabel(index, label)
			}(index)
		}

	default:
		logger.Debugf("Got notification %s %+v", reflect.TypeOf(cmd), cmd)
	}
}
