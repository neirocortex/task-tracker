package grpc

import (
	"context"
	"taskTracker/internal/domain"
	"taskTracker/internal/usecase"

	taskv1 "taskTracker/internal/delivery/grpc/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TagHandler struct {
	taskv1.UnimplementedTagServiceServer

	createCmd *usecase.CreateTagCommand
	updateCmd *usecase.UpdateTagCommand
	deleteCmd *usecase.DeleteTagCommand
	listQuery *usecase.GetTagsQuery
}

func NewTagHandler(
	createCmd *usecase.CreateTagCommand,
	updateCmd *usecase.UpdateTagCommand,
	deleteCmd *usecase.DeleteTagCommand,
	listQuery *usecase.GetTagsQuery,
) *TagHandler {
	return &TagHandler{
		createCmd: createCmd,
		updateCmd: updateCmd,
		deleteCmd: deleteCmd,
		listQuery: listQuery,
	}
}

func mapTagDomainErrorToGRPC(err error) error {
	switch err {
	case domain.ErrTagInvalid:
		return status.Error(codes.InvalidArgument, err.Error())
	case domain.ErrSystemTagModification:
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

func toPbTag(tag *domain.Tag) *taskv1.TagResponse {
	if tag == nil {
		return nil
	}

	return &taskv1.TagResponse{
		Name:     tag.Name,
		IsSystem: tag.IsSystem,
	}
}

func (h *TagHandler) CreateTag(ctx context.Context, req *taskv1.CreateTagRequest) (*taskv1.CreateTagResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	err := h.createCmd.Execute(ctx, req.Name)
	if err != nil {
		return nil, mapTagDomainErrorToGRPC(err)
	}

	return &taskv1.CreateTagResponse{}, nil
}
func (h *TagHandler) UpdateTag(ctx context.Context, req *taskv1.UpdateTagRequest) (*taskv1.UpdateTagResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	err := h.updateCmd.Execute(ctx, req.Name, req.NewName)
	if err != nil {
		return nil, mapTagDomainErrorToGRPC(err)
	}

	return &taskv1.UpdateTagResponse{}, nil
}
func (h *TagHandler) DeleteTag(ctx context.Context, req *taskv1.DeleteTagRequest) (*taskv1.DeleteTagResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	err := h.deleteCmd.Execute(ctx, req.Name)
	if err != nil {
		return nil, mapTagDomainErrorToGRPC(err)
	}

	return &taskv1.DeleteTagResponse{}, nil
}
func (h *TagHandler) GetTags(ctx context.Context, req *taskv1.GetTagsRequest) (*taskv1.GetTagsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	paginatedData, err := h.listQuery.Execute(ctx, int(req.Limit), int(req.Page))
	if err != nil {
		return nil, mapTagDomainErrorToGRPC(err)
	}

	tagsPb := make([]*taskv1.TagResponse, len(paginatedData.Tags))
	for i := range paginatedData.Tags {
		tagsPb[i] = toPbTag(&paginatedData.Tags[i])
	}

	response := &taskv1.GetTagsResponse{
		Data: tagsPb,
		Pagination: &taskv1.TagPaginationMeta{
			CurrentPage: req.Page,
			Limit:       req.Limit,
			TotalItems:  int64(paginatedData.TotalCount),
			TotalPages:  int64(paginatedData.TotalPages),
		},
	}
	return response, nil
}
