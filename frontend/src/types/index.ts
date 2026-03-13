// ==================== 用户与认证 ====================

export interface User {
    id: number;
    username: string;
    email: string;
    role: string;
    roleCode: string;
    status: number;
    lastLogin?: string;
}

export interface LoginRequest {
    username: string;
    password: string;
}

export interface LoginResponse {
    token: string;
    expiresAt: string;
    user: User;
}

export interface AuthStore {
    isAuthenticated: boolean;
    user: User | null;
    token: string | null;
}

// ==================== PCAP 文件 ====================

export interface PcapFile {
    id: number;
    fileId: string;
    fileName: string;
    filePath: string;
    fileSize: string;
    status: string;
    createdAt: string;
    _progress?: number;
    _error?: string;
}

export interface FileQueryReq {
    fileName?: string;
    fileSize?: string;
    startDate?: string;
    endDate?: string;
    page: number;
    pageSize: number;
}

export interface FileQueryResp {
    list: PcapFile[];
    total: number;
}

// ==================== 分析器 ====================

export interface Packet {
    index: number;
    timestamp: string;
    source: string;
    dest: string;
    protocol: string;
    length: number;
    info: string;
}

export interface PacketDetail {
    protocolTree: any;
    hexData: string;
}

export interface StreamData {
    payloads: Array<{
        dir: 'client' | 'server';
        hexData: string;
    }>;
    clientBytes: number;
    serverBytes: number;
    packetCount: number;
}

// ==================== 网络接口 ====================

export interface NetworkInterface {
    name: string;
    description?: string;
    addresses?: Array<{
        ip: string;
        netmask: string;
    }>;
    status: string;
}

// ==================== 通用 ====================

export interface ApiResponse<T = any> {
    code: number;
    msg: string;
    error?: string;
    data: T;
}

export interface ProgressEvent {
    fileId: string;
    fileName?: string;
    percent: number;
    error?: string;
}
