// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.2
// source: filetransfer.proto

package filetransfer

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// look up in mongo db if the file exists
// the file change should be the first it of information that is sent
// if the file does not exist yeah, then send the entire file
// TODO: add the directory watching to the grpc
type FileData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`                                 // Persistent unique identifier for the file
	Content   []byte `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`                       // File content (for upload/download)
	Location  string `protobuf:"bytes,3,opt,name=location,proto3" json:"location,omitempty"`                     // location of the file
	Offset    int64  `protobuf:"varint,4,opt,name=offset,proto3" json:"offset,omitempty"`                        // Offset for streaming
	TotalSize int64  `protobuf:"varint,5,opt,name=total_size,json=totalSize,proto3" json:"total_size,omitempty"` // Total size of the file
}

func (x *FileData) Reset() {
	*x = FileData{}
	mi := &file_filetransfer_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FileData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileData) ProtoMessage() {}

func (x *FileData) ProtoReflect() protoreflect.Message {
	mi := &file_filetransfer_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileData.ProtoReflect.Descriptor instead.
func (*FileData) Descriptor() ([]byte, []int) {
	return file_filetransfer_proto_rawDescGZIP(), []int{0}
}

func (x *FileData) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *FileData) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

func (x *FileData) GetLocation() string {
	if x != nil {
		return x.Location
	}
	return ""
}

func (x *FileData) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *FileData) GetTotalSize() int64 {
	if x != nil {
		return x.TotalSize
	}
	return 0
}

type FileChange struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content   string `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`                       // File content
	Location  string `protobuf:"bytes,2,opt,name=location,proto3" json:"location,omitempty"`                     // Path to the file
	FileId    string `protobuf:"bytes,3,opt,name=file_id,json=fileId,proto3" json:"file_id,omitempty"`           // Unique identifier for the file
	ChangeId  string `protobuf:"bytes,4,opt,name=change_id,json=changeId,proto3" json:"change_id,omitempty"`     // Unique identifier for the change
	Offset    int64  `protobuf:"varint,5,opt,name=offset,proto3" json:"offset,omitempty"`                        // Offset for streaming
	TotalSize int64  `protobuf:"varint,6,opt,name=total_size,json=totalSize,proto3" json:"total_size,omitempty"` // Total size of the file change
	Timestamp int64  `protobuf:"varint,7,opt,name=timestamp,proto3" json:"timestamp,omitempty"`                  // Timestamp for the change
}

func (x *FileChange) Reset() {
	*x = FileChange{}
	mi := &file_filetransfer_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FileChange) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileChange) ProtoMessage() {}

func (x *FileChange) ProtoReflect() protoreflect.Message {
	mi := &file_filetransfer_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileChange.ProtoReflect.Descriptor instead.
func (*FileChange) Descriptor() ([]byte, []int) {
	return file_filetransfer_proto_rawDescGZIP(), []int{1}
}

func (x *FileChange) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *FileChange) GetLocation() string {
	if x != nil {
		return x.Location
	}
	return ""
}

func (x *FileChange) GetFileId() string {
	if x != nil {
		return x.FileId
	}
	return ""
}

func (x *FileChange) GetChangeId() string {
	if x != nil {
		return x.ChangeId
	}
	return ""
}

func (x *FileChange) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *FileChange) GetTotalSize() int64 {
	if x != nil {
		return x.TotalSize
	}
	return 0
}

func (x *FileChange) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type FileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`     // Persistent unique identifier for the file
	Path string `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"` // Path to the file (used for delete or download)
}

func (x *FileRequest) Reset() {
	*x = FileRequest{}
	mi := &file_filetransfer_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileRequest) ProtoMessage() {}

func (x *FileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_filetransfer_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileRequest.ProtoReflect.Descriptor instead.
func (*FileRequest) Descriptor() ([]byte, []int) {
	return file_filetransfer_proto_rawDescGZIP(), []int{2}
}

func (x *FileRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *FileRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

type MoveFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`                   // Persistent unique identifier for the file
	Destination string `protobuf:"bytes,2,opt,name=destination,proto3" json:"destination,omitempty"` // New path after move/rename
}

func (x *MoveFileRequest) Reset() {
	*x = MoveFileRequest{}
	mi := &file_filetransfer_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MoveFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MoveFileRequest) ProtoMessage() {}

func (x *MoveFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_filetransfer_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MoveFileRequest.ProtoReflect.Descriptor instead.
func (*MoveFileRequest) Descriptor() ([]byte, []int) {
	return file_filetransfer_proto_rawDescGZIP(), []int{3}
}

func (x *MoveFileRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *MoveFileRequest) GetDestination() string {
	if x != nil {
		return x.Destination
	}
	return ""
}

type ActionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ActionResponse) Reset() {
	*x = ActionResponse{}
	mi := &file_filetransfer_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ActionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ActionResponse) ProtoMessage() {}

func (x *ActionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_filetransfer_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ActionResponse.ProtoReflect.Descriptor instead.
func (*ActionResponse) Descriptor() ([]byte, []int) {
	return file_filetransfer_proto_rawDescGZIP(), []int{4}
}

func (x *ActionResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *ActionResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_filetransfer_proto protoreflect.FileDescriptor

var file_filetransfer_proto_rawDesc = []byte{
	0x0a, 0x12, 0x66, 0x69, 0x6c, 0x65, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x66, 0x69, 0x6c, 0x65, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66,
	0x65, 0x72, 0x22, 0x87, 0x01, 0x0a, 0x08, 0x46, 0x69, 0x6c, 0x65, 0x44, 0x61, 0x74, 0x61, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x1d, 0x0a,
	0x0a, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x09, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x53, 0x69, 0x7a, 0x65, 0x22, 0xcd, 0x01, 0x0a,
	0x0a, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x17, 0x0a, 0x07, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x66, 0x69, 0x6c, 0x65, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x63, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63,
	0x68, 0x61, 0x6e, 0x67, 0x65, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65,
	0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12,
	0x1d, 0x0a, 0x0a, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x1c,
	0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x31, 0x0a, 0x0b,
	0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x70,
	0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x22,
	0x43, 0x0a, 0x0f, 0x4d, 0x6f, 0x76, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x22, 0x44, 0x0a, 0x0e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0xad, 0x02, 0x0a, 0x0b, 0x46,
	0x69, 0x6c, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x41, 0x0a, 0x0b, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x16, 0x2e, 0x66, 0x69, 0x6c, 0x65,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x44, 0x61, 0x74,
	0x61, 0x1a, 0x16, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72,
	0x2e, 0x46, 0x69, 0x6c, 0x65, 0x44, 0x61, 0x74, 0x61, 0x28, 0x01, 0x30, 0x01, 0x12, 0x4b, 0x0a,
	0x11, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x73, 0x12, 0x18, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65,
	0x72, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x1a, 0x18, 0x2e, 0x66,
	0x69, 0x6c, 0x65, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x46, 0x69, 0x6c, 0x65,
	0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x28, 0x01, 0x30, 0x01, 0x12, 0x45, 0x0a, 0x0a, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x19, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66,
	0x65, 0x72, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x47, 0x0a, 0x08, 0x4d, 0x6f, 0x76, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x1d, 0x2e,
	0x66, 0x69, 0x6c, 0x65, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x4d, 0x6f, 0x76,
	0x65, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x66,
	0x69, 0x6c, 0x65, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x2e, 0x41, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x11, 0x5a, 0x0f, 0x2e, 0x2e,
	0x2f, 0x66, 0x69, 0x6c, 0x65, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_filetransfer_proto_rawDescOnce sync.Once
	file_filetransfer_proto_rawDescData = file_filetransfer_proto_rawDesc
)

func file_filetransfer_proto_rawDescGZIP() []byte {
	file_filetransfer_proto_rawDescOnce.Do(func() {
		file_filetransfer_proto_rawDescData = protoimpl.X.CompressGZIP(file_filetransfer_proto_rawDescData)
	})
	return file_filetransfer_proto_rawDescData
}

var file_filetransfer_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_filetransfer_proto_goTypes = []any{
	(*FileData)(nil),        // 0: filetransfer.FileData
	(*FileChange)(nil),      // 1: filetransfer.FileChange
	(*FileRequest)(nil),     // 2: filetransfer.FileRequest
	(*MoveFileRequest)(nil), // 3: filetransfer.MoveFileRequest
	(*ActionResponse)(nil),  // 4: filetransfer.ActionResponse
}
var file_filetransfer_proto_depIdxs = []int32{
	0, // 0: filetransfer.FileService.StreamFiles:input_type -> filetransfer.FileData
	1, // 1: filetransfer.FileService.StreamFileChanges:input_type -> filetransfer.FileChange
	2, // 2: filetransfer.FileService.DeleteFile:input_type -> filetransfer.FileRequest
	3, // 3: filetransfer.FileService.MoveFile:input_type -> filetransfer.MoveFileRequest
	0, // 4: filetransfer.FileService.StreamFiles:output_type -> filetransfer.FileData
	1, // 5: filetransfer.FileService.StreamFileChanges:output_type -> filetransfer.FileChange
	4, // 6: filetransfer.FileService.DeleteFile:output_type -> filetransfer.ActionResponse
	4, // 7: filetransfer.FileService.MoveFile:output_type -> filetransfer.ActionResponse
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_filetransfer_proto_init() }
func file_filetransfer_proto_init() {
	if File_filetransfer_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_filetransfer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_filetransfer_proto_goTypes,
		DependencyIndexes: file_filetransfer_proto_depIdxs,
		MessageInfos:      file_filetransfer_proto_msgTypes,
	}.Build()
	File_filetransfer_proto = out.File
	file_filetransfer_proto_rawDesc = nil
	file_filetransfer_proto_goTypes = nil
	file_filetransfer_proto_depIdxs = nil
}
