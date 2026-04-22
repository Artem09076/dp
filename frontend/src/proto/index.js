import * as grpcWeb from './auth_grpc_web_pb'
import * as authPb from './auth_pb'

export const AuthServiceClient = grpcWeb.AuthServiceClient
export const RegisterRequest = authPb.RegisterRequest
export const LoginRequest = authPb.LoginRequest
export const RefreshTokenRequest = authPb.RefreshTokenRequest
export const LogoutRequest = authPb.LogoutRequest