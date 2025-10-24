package handler

import (
	"context"

	"ironnode/pkg/models"
	"ironnode/services/blockchain-service/internal/service"
	pb "ironnode/services/blockchain-service/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NodeHandler struct {
	pb.UnimplementedBlockchainServiceServer
	nodeService service.NodeService
}

func NewNodeHandler(nodeService service.NodeService) *NodeHandler {
	return &NodeHandler{nodeService: nodeService}
}

func (h *NodeHandler) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.NodeResponse, error) {
	blockchainType := models.BlockchainType(req.Type)

	node, err := h.nodeService.CreateNode(req.Name, blockchainType, req.Network, req.Url)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create node: %v", err)
	}

	return &pb.NodeResponse{
		Id:       node.ID.String(),
		Name:     node.Name,
		Type:     string(node.Type),
		Network:  node.Network,
		Url:      node.URL,
		IsActive: node.IsActive,
		Priority: int32(node.Priority),
	}, nil
}

func (h *NodeHandler) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.NodeResponse, error) {
	nodeID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid node ID: %v", err)
	}

	node, err := h.nodeService.GetNodeByID(nodeID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "node not found: %v", err)
	}

	return &pb.NodeResponse{
		Id:       node.ID.String(),
		Name:     node.Name,
		Type:     string(node.Type),
		Network:  node.Network,
		Url:      node.URL,
		IsActive: node.IsActive,
		Priority: int32(node.Priority),
	}, nil
}

func (h *NodeHandler) ListNodes(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	nodes, err := h.nodeService.GetActiveNodes()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list nodes: %v", err)
	}

	var pbNodes []*pb.NodeResponse
	for _, node := range nodes {
		pbNodes = append(pbNodes, &pb.NodeResponse{
			Id:       node.ID.String(),
			Name:     node.Name,
			Type:     string(node.Type),
			Network:  node.Network,
			Url:      node.URL,
			IsActive: node.IsActive,
			Priority: int32(node.Priority),
		})
	}

	return &pb.ListNodesResponse{
		Nodes: pbNodes,
	}, nil
}

func (h *NodeHandler) GetNodesByType(ctx context.Context, req *pb.GetNodesByTypeRequest) (*pb.ListNodesResponse, error) {
	blockchainType := models.BlockchainType(req.Type)

	nodes, err := h.nodeService.GetNodesByType(blockchainType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get nodes by type: %v", err)
	}

	var pbNodes []*pb.NodeResponse
	for _, node := range nodes {
		pbNodes = append(pbNodes, &pb.NodeResponse{
			Id:       node.ID.String(),
			Name:     node.Name,
			Type:     string(node.Type),
			Network:  node.Network,
			Url:      node.URL,
			IsActive: node.IsActive,
			Priority: int32(node.Priority),
		})
	}

	return &pb.ListNodesResponse{
		Nodes: pbNodes,
	}, nil
}
