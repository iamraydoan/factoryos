package factoryos.resource.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.64.0)",
    comments = "Source: resource/v1/equipment.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class EquipmentServiceGrpc {

  private EquipmentServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "factoryos.resource.v1.EquipmentService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.CreateWorkUnitRequest,
      factoryos.resource.v1.Equipment.CreateWorkUnitResponse> getCreateWorkUnitMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateWorkUnit",
      requestType = factoryos.resource.v1.Equipment.CreateWorkUnitRequest.class,
      responseType = factoryos.resource.v1.Equipment.CreateWorkUnitResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.CreateWorkUnitRequest,
      factoryos.resource.v1.Equipment.CreateWorkUnitResponse> getCreateWorkUnitMethod() {
    io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.CreateWorkUnitRequest, factoryos.resource.v1.Equipment.CreateWorkUnitResponse> getCreateWorkUnitMethod;
    if ((getCreateWorkUnitMethod = EquipmentServiceGrpc.getCreateWorkUnitMethod) == null) {
      synchronized (EquipmentServiceGrpc.class) {
        if ((getCreateWorkUnitMethod = EquipmentServiceGrpc.getCreateWorkUnitMethod) == null) {
          EquipmentServiceGrpc.getCreateWorkUnitMethod = getCreateWorkUnitMethod =
              io.grpc.MethodDescriptor.<factoryos.resource.v1.Equipment.CreateWorkUnitRequest, factoryos.resource.v1.Equipment.CreateWorkUnitResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateWorkUnit"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  factoryos.resource.v1.Equipment.CreateWorkUnitRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  factoryos.resource.v1.Equipment.CreateWorkUnitResponse.getDefaultInstance()))
              .setSchemaDescriptor(new EquipmentServiceMethodDescriptorSupplier("CreateWorkUnit"))
              .build();
        }
      }
    }
    return getCreateWorkUnitMethod;
  }

  private static volatile io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.GetWorkUnitRequest,
      factoryos.resource.v1.Equipment.GetWorkUnitResponse> getGetWorkUnitMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetWorkUnit",
      requestType = factoryos.resource.v1.Equipment.GetWorkUnitRequest.class,
      responseType = factoryos.resource.v1.Equipment.GetWorkUnitResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.GetWorkUnitRequest,
      factoryos.resource.v1.Equipment.GetWorkUnitResponse> getGetWorkUnitMethod() {
    io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.GetWorkUnitRequest, factoryos.resource.v1.Equipment.GetWorkUnitResponse> getGetWorkUnitMethod;
    if ((getGetWorkUnitMethod = EquipmentServiceGrpc.getGetWorkUnitMethod) == null) {
      synchronized (EquipmentServiceGrpc.class) {
        if ((getGetWorkUnitMethod = EquipmentServiceGrpc.getGetWorkUnitMethod) == null) {
          EquipmentServiceGrpc.getGetWorkUnitMethod = getGetWorkUnitMethod =
              io.grpc.MethodDescriptor.<factoryos.resource.v1.Equipment.GetWorkUnitRequest, factoryos.resource.v1.Equipment.GetWorkUnitResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetWorkUnit"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  factoryos.resource.v1.Equipment.GetWorkUnitRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  factoryos.resource.v1.Equipment.GetWorkUnitResponse.getDefaultInstance()))
              .setSchemaDescriptor(new EquipmentServiceMethodDescriptorSupplier("GetWorkUnit"))
              .build();
        }
      }
    }
    return getGetWorkUnitMethod;
  }

  private static volatile io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.ListWorkUnitsRequest,
      factoryos.resource.v1.Equipment.ListWorkUnitsResponse> getListWorkUnitsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ListWorkUnits",
      requestType = factoryos.resource.v1.Equipment.ListWorkUnitsRequest.class,
      responseType = factoryos.resource.v1.Equipment.ListWorkUnitsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.ListWorkUnitsRequest,
      factoryos.resource.v1.Equipment.ListWorkUnitsResponse> getListWorkUnitsMethod() {
    io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.ListWorkUnitsRequest, factoryos.resource.v1.Equipment.ListWorkUnitsResponse> getListWorkUnitsMethod;
    if ((getListWorkUnitsMethod = EquipmentServiceGrpc.getListWorkUnitsMethod) == null) {
      synchronized (EquipmentServiceGrpc.class) {
        if ((getListWorkUnitsMethod = EquipmentServiceGrpc.getListWorkUnitsMethod) == null) {
          EquipmentServiceGrpc.getListWorkUnitsMethod = getListWorkUnitsMethod =
              io.grpc.MethodDescriptor.<factoryos.resource.v1.Equipment.ListWorkUnitsRequest, factoryos.resource.v1.Equipment.ListWorkUnitsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ListWorkUnits"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  factoryos.resource.v1.Equipment.ListWorkUnitsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  factoryos.resource.v1.Equipment.ListWorkUnitsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new EquipmentServiceMethodDescriptorSupplier("ListWorkUnits"))
              .build();
        }
      }
    }
    return getListWorkUnitsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest,
      factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse> getUpdateWorkUnitStatusMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateWorkUnitStatus",
      requestType = factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest.class,
      responseType = factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest,
      factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse> getUpdateWorkUnitStatusMethod() {
    io.grpc.MethodDescriptor<factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest, factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse> getUpdateWorkUnitStatusMethod;
    if ((getUpdateWorkUnitStatusMethod = EquipmentServiceGrpc.getUpdateWorkUnitStatusMethod) == null) {
      synchronized (EquipmentServiceGrpc.class) {
        if ((getUpdateWorkUnitStatusMethod = EquipmentServiceGrpc.getUpdateWorkUnitStatusMethod) == null) {
          EquipmentServiceGrpc.getUpdateWorkUnitStatusMethod = getUpdateWorkUnitStatusMethod =
              io.grpc.MethodDescriptor.<factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest, factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateWorkUnitStatus"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse.getDefaultInstance()))
              .setSchemaDescriptor(new EquipmentServiceMethodDescriptorSupplier("UpdateWorkUnitStatus"))
              .build();
        }
      }
    }
    return getUpdateWorkUnitStatusMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static EquipmentServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<EquipmentServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<EquipmentServiceStub>() {
        @java.lang.Override
        public EquipmentServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new EquipmentServiceStub(channel, callOptions);
        }
      };
    return EquipmentServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static EquipmentServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<EquipmentServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<EquipmentServiceBlockingStub>() {
        @java.lang.Override
        public EquipmentServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new EquipmentServiceBlockingStub(channel, callOptions);
        }
      };
    return EquipmentServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static EquipmentServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<EquipmentServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<EquipmentServiceFutureStub>() {
        @java.lang.Override
        public EquipmentServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new EquipmentServiceFutureStub(channel, callOptions);
        }
      };
    return EquipmentServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     * <pre>
     * Work Unit API
     * </pre>
     */
    default void createWorkUnit(factoryos.resource.v1.Equipment.CreateWorkUnitRequest request,
        io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.CreateWorkUnitResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateWorkUnitMethod(), responseObserver);
    }

    /**
     */
    default void getWorkUnit(factoryos.resource.v1.Equipment.GetWorkUnitRequest request,
        io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.GetWorkUnitResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetWorkUnitMethod(), responseObserver);
    }

    /**
     */
    default void listWorkUnits(factoryos.resource.v1.Equipment.ListWorkUnitsRequest request,
        io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.ListWorkUnitsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListWorkUnitsMethod(), responseObserver);
    }

    /**
     */
    default void updateWorkUnitStatus(factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest request,
        io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateWorkUnitStatusMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service EquipmentService.
   */
  public static abstract class EquipmentServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return EquipmentServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service EquipmentService.
   */
  public static final class EquipmentServiceStub
      extends io.grpc.stub.AbstractAsyncStub<EquipmentServiceStub> {
    private EquipmentServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected EquipmentServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new EquipmentServiceStub(channel, callOptions);
    }

    /**
     * <pre>
     * Work Unit API
     * </pre>
     */
    public void createWorkUnit(factoryos.resource.v1.Equipment.CreateWorkUnitRequest request,
        io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.CreateWorkUnitResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateWorkUnitMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void getWorkUnit(factoryos.resource.v1.Equipment.GetWorkUnitRequest request,
        io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.GetWorkUnitResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetWorkUnitMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void listWorkUnits(factoryos.resource.v1.Equipment.ListWorkUnitsRequest request,
        io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.ListWorkUnitsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListWorkUnitsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateWorkUnitStatus(factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest request,
        io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateWorkUnitStatusMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service EquipmentService.
   */
  public static final class EquipmentServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<EquipmentServiceBlockingStub> {
    private EquipmentServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected EquipmentServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new EquipmentServiceBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * Work Unit API
     * </pre>
     */
    public factoryos.resource.v1.Equipment.CreateWorkUnitResponse createWorkUnit(factoryos.resource.v1.Equipment.CreateWorkUnitRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateWorkUnitMethod(), getCallOptions(), request);
    }

    /**
     */
    public factoryos.resource.v1.Equipment.GetWorkUnitResponse getWorkUnit(factoryos.resource.v1.Equipment.GetWorkUnitRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetWorkUnitMethod(), getCallOptions(), request);
    }

    /**
     */
    public factoryos.resource.v1.Equipment.ListWorkUnitsResponse listWorkUnits(factoryos.resource.v1.Equipment.ListWorkUnitsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListWorkUnitsMethod(), getCallOptions(), request);
    }

    /**
     */
    public factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse updateWorkUnitStatus(factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateWorkUnitStatusMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service EquipmentService.
   */
  public static final class EquipmentServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<EquipmentServiceFutureStub> {
    private EquipmentServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected EquipmentServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new EquipmentServiceFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * Work Unit API
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<factoryos.resource.v1.Equipment.CreateWorkUnitResponse> createWorkUnit(
        factoryos.resource.v1.Equipment.CreateWorkUnitRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateWorkUnitMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<factoryos.resource.v1.Equipment.GetWorkUnitResponse> getWorkUnit(
        factoryos.resource.v1.Equipment.GetWorkUnitRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetWorkUnitMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<factoryos.resource.v1.Equipment.ListWorkUnitsResponse> listWorkUnits(
        factoryos.resource.v1.Equipment.ListWorkUnitsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListWorkUnitsMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse> updateWorkUnitStatus(
        factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateWorkUnitStatusMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_CREATE_WORK_UNIT = 0;
  private static final int METHODID_GET_WORK_UNIT = 1;
  private static final int METHODID_LIST_WORK_UNITS = 2;
  private static final int METHODID_UPDATE_WORK_UNIT_STATUS = 3;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final AsyncService serviceImpl;
    private final int methodId;

    MethodHandlers(AsyncService serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_CREATE_WORK_UNIT:
          serviceImpl.createWorkUnit((factoryos.resource.v1.Equipment.CreateWorkUnitRequest) request,
              (io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.CreateWorkUnitResponse>) responseObserver);
          break;
        case METHODID_GET_WORK_UNIT:
          serviceImpl.getWorkUnit((factoryos.resource.v1.Equipment.GetWorkUnitRequest) request,
              (io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.GetWorkUnitResponse>) responseObserver);
          break;
        case METHODID_LIST_WORK_UNITS:
          serviceImpl.listWorkUnits((factoryos.resource.v1.Equipment.ListWorkUnitsRequest) request,
              (io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.ListWorkUnitsResponse>) responseObserver);
          break;
        case METHODID_UPDATE_WORK_UNIT_STATUS:
          serviceImpl.updateWorkUnitStatus((factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest) request,
              (io.grpc.stub.StreamObserver<factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getCreateWorkUnitMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              factoryos.resource.v1.Equipment.CreateWorkUnitRequest,
              factoryos.resource.v1.Equipment.CreateWorkUnitResponse>(
                service, METHODID_CREATE_WORK_UNIT)))
        .addMethod(
          getGetWorkUnitMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              factoryos.resource.v1.Equipment.GetWorkUnitRequest,
              factoryos.resource.v1.Equipment.GetWorkUnitResponse>(
                service, METHODID_GET_WORK_UNIT)))
        .addMethod(
          getListWorkUnitsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              factoryos.resource.v1.Equipment.ListWorkUnitsRequest,
              factoryos.resource.v1.Equipment.ListWorkUnitsResponse>(
                service, METHODID_LIST_WORK_UNITS)))
        .addMethod(
          getUpdateWorkUnitStatusMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              factoryos.resource.v1.Equipment.UpdateWorkUnitStatusRequest,
              factoryos.resource.v1.Equipment.UpdateWorkUnitStatusResponse>(
                service, METHODID_UPDATE_WORK_UNIT_STATUS)))
        .build();
  }

  private static abstract class EquipmentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    EquipmentServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return factoryos.resource.v1.Equipment.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("EquipmentService");
    }
  }

  private static final class EquipmentServiceFileDescriptorSupplier
      extends EquipmentServiceBaseDescriptorSupplier {
    EquipmentServiceFileDescriptorSupplier() {}
  }

  private static final class EquipmentServiceMethodDescriptorSupplier
      extends EquipmentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    EquipmentServiceMethodDescriptorSupplier(java.lang.String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (EquipmentServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new EquipmentServiceFileDescriptorSupplier())
              .addMethod(getCreateWorkUnitMethod())
              .addMethod(getGetWorkUnitMethod())
              .addMethod(getListWorkUnitsMethod())
              .addMethod(getUpdateWorkUnitStatusMethod())
              .build();
        }
      }
    }
    return result;
  }
}
