export interface Task {
  [key: string]: any;
  id: string;
  title: string;
  assignee_id: string | null;
  status: string;
  assignee: string; // 💡 전역 타입 호환성 유지
}
export interface Column {
  id: string;
  title: string;
  tasks: Task[];
}

export interface TaskComment {
  id: number;
  author: string;
  content: string;
  timestamp: string;
}

export interface UserProfile {
  profileId: string;
  userId: string;
  name: string;
  email: string | null;
  profileImageUrl: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface User {
  id: number;
  email: string;
  name: string;
  profileImage?: string;
}

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
}

export interface Workspace {
  id: number;
  name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Project {
  id: number;
  workspaceId: number;
  name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Comment {
  id: number;
  taskId: number;
  userId: number;
  content: string;
  createdAt: string;
  user?: User;
}

export interface Board {
  id: number;
  title: string;
  description?: string;
  ownerId: number;
  createdAt: string;
  updatedAt: string;
}
