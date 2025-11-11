// src/types/board.ts

// Custom Field API 응답에서 필요한 기본 구조를 정의합니다.
// Stage/Role/Importance는 Field/Option으로 대체됩니다.

/**
 * 프로젝트 커스텀 필드의 응답 타입
 */
export interface FieldResponse {
  fieldId: string;
  projectId: string;
  name: string;
  description: string;
  fieldType:
    | 'text'
    | 'number'
    | 'single_select'
    | 'multi_select'
    | 'date'
    | 'datetime'
    | 'single_user'
    | 'multi_user'
    | 'checkbox'
    | 'url';
  isRequired: boolean;
  config: Record<string, any>;
}

/**
 * 커스텀 필드 옵션 (예: Stage, Role, Importance의 선택지)의 응답 타입
 */
export interface FieldOptionResponse {
  // optionId: string; // 💡 이 속성이 필수!
  fieldId: string;
  label: string;
  description: string;
  color: string;
  displayOrder: number;
}

export interface BaseFieldOption {
  fieldId: string;
  label: string;
  description: string;
  color: string;
  displayOrder: number;
  isSystemDefault?: boolean;
}

/**
 * Stage 옵션 타입 (Stage ID를 옵션의 고유 ID로 사용)
 */
export interface CustomStageResponse extends BaseFieldOption {
  stageId: string; // 고유 ID 역할
}

/**
 * Role 옵션 타입
 */
export interface CustomRoleResponse extends BaseFieldOption {
  roleId: string; // 고유 ID 역할
}

/**
 * Importance 옵션 타입
 */
export interface CustomImportanceResponse extends BaseFieldOption {
  importanceId: string; // 고유 ID 역할
  level?: number;
}

// =======================================================
// Board API 요청/응답 타입 (boardService.ts에서 사용하는 타입과 동일)
// =======================================================

/**
 * 보드 생성 요청 (POST /boards)
 */
export interface CreateBoardRequest {
  projectId: string;
  title: string;
  content?: string;
  assigneeId?: string;
  dueDate?: string;
  stageId?: string;
  importanceId?: string;
  roleIds?: string[];
}

/**
 * 보드 수정 요청 (PUT /boards/{boardId})
 */
export interface UpdateBoardRequest extends Partial<CreateBoardRequest> {}

/**
 * 보드 응답 (GET /boards/{boardId})
 */
export interface BoardResponse {
  boardId: string;
  title: string;
  content: string;
  projectId: string;
  position: string;
  dueDate: string;
  createdAt: string;
  updatedAt: string;
  author: {
    userId: string;
    name: string;
    email: string;
    isActive: boolean;
  };
  assignee: {
    userId: string;
    name: string;
    email: string;
    isActive: boolean;
  };
  customFields: Record<string, any>;
  stage?: { stageId: string; name: string; color: string };
}

// =======================================================
// 기존 프론트 컴포넌트에서 요청하신 타입 (Kanban -> Board)
// =======================================================

export interface CustomField {
  id: string;
  name: string;
  type: 'TEXT' | 'NUMBER' | 'DATE' | 'PERSON' | 'SELECT';
  options?: { value: string; isDefault: boolean }[];
  allowMultipleSections?: boolean;
  defaultValue?: any;
}

export type Priority = 'HIGH' | 'MEDIUM' | 'LOW' | '';

export interface Board {
  id: string;
  title: string;
  assigneeId: string;
  status: string;
  assignee: string;
  description?: string;
  dueDate?: string;
  priority?: Priority;
}

export interface BoardWithCustomFields extends Board {
  customFieldValues?: {
    [key: string]: any;
  };
}
