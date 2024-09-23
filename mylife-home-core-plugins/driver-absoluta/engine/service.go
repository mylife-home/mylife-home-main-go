package engine

import (
	"mylife-home-core-plugins-driver-absoluta/engine/itv2"
	"mylife-home-core-plugins-driver-absoluta/engine/itv2/commands"
	"time"
)

type Service struct {
	client           *itv2.Client
	state            *stateUpdater
	connectedChanged func(bool)
}

func NewService(serverAddress string, uid string, pin string, state State, connectedChanged func(bool)) *Service {

	svc := &Service{
		client:           itv2.MakeClient(serverAddress, uid, pin),
		state:            makeStateUpdater(state),
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
			logger.WithError(err).Debug(" End GetPartitionStatus")
			if client.Status() != itv2.ConnectionOpen {
				return
			}

			logger.WithError(err).Error("Error reading partition statuses")
		} else {
			state.UpdatePartitionStatuses(statuses.GetData())
		}

		if statuses, err := client.GetZoneStatuses(); err != nil {
			logger.WithError(err).Debug(" End GetZoneStatuses")
			if client.Status() != itv2.ConnectionOpen {
				return
			}

			logger.WithError(err).Error("Error reading zone statuses")
		} else {
			state.UpdateZoneStatuses(statuses.GetData())
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
		state.UpdateCapabilities(partitionCount, zoneCount)

	case *commands.PartitionAssignmentConfiguration:
		assignedPartitions := cmd.GetAssignedPartitions()
		logger.Debugf("Got PartitionAssignmentConfiguration: %+v", assignedPartitions)
		state.UpdateAssignedPartitions(assignedPartitions)

		for _, index := range assignedPartitions {
			go func(index int) {
				label, err := client.GetPartitionLabel(index)
				if err != nil {
					logger.WithError(err).Errorf("Error reading partition label %d", index)
					return
				}

				state.UpdatePartitionLabel(index, label)
			}(index)
		}

	case *commands.ZoneAssignmentConfiguration:
		assignedZones := cmd.GetAssignedZones()
		// Note: cmd.Req.PartitionNumber.GetUint() seems unused (always 0)
		logger.Debugf("Got ZoneAssignmentConfiguration: %+v", assignedZones)
		state.UpdateAssignedZones(assignedZones)

		for _, index := range assignedZones {
			go func(index int) {
				label, err := client.GetZoneLabel(index)
				if err != nil {
					logger.WithError(err).Errorf("Error reading zone label %d", index)
					return
				}

				state.UpdateZoneLabel(index, label)
			}(index)
		}

	default:
		// logger.Debugf("Got notification %s %+v", reflect.TypeOf(cmd), cmd)
	}
}
