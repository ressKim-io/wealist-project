// (신규 또는 이름 변경) src/types/board.ts

// 💡 CustomField 타입은 그대로 유지합니다.
export interface CustomField {
  id: string;
  name: string;
  type: 'TEXT' | 'NUMBER' | 'DATE' | 'PERSON' | 'SELECT';
  options?: { value: string; isDefault: boolean }[];
  allowMultipleSections?: boolean;
  defaultValue?: any;
}

export type Priority = 'HIGH' | 'MEDIUM' | 'LOW' | '';

// 1. 'Kanban' -> 'Board'
export interface Board {
  id: string;
  title: string;
  assignee_id: string;
  status: string;
  assignee: string;
  description?: string;
  dueDate?: string;
  priority?: Priority;
}

// 2. 'KanbanWithCustomFields' -> 'BoardWithCustomFields'
export interface BoardWithCustomFields extends Board {
  customFieldValues?: {
    [key: string]: any; // (타입을 any로 변경하여 유연성 확보)
  };
}
