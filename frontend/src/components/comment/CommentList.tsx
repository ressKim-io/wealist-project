// src/components/board/CommentList.tsx

import React, { useState, useEffect, useRef } from 'react';
import { Pencil, Trash2, Check, X, Send, Paperclip } from 'lucide-react';

import { CommentResponse } from '../../types/board';
import {
  deleteComment,
  updateComment,
  createComment,
  uploadAttachment,
} from '../../api/board/boardService'; // api -> apis 경로 확인

import { WorkspaceMemberResponse } from '../../types/user';
import { useUserLookup } from '../../hooks/useUserLookup';
import { useFileUpload } from '../../hooks/useFileUpload';

// =============================================================================
// [Sub Component] 댓글 작성 인풋 (Compact Style)
// =============================================================================
interface CommentInputProps {
  boardId: string;
  workspaceId: string;
  currentUserId: string;
  onCommentCreated: () => void;
}

const CommentInput = ({
  boardId,
  workspaceId,
  currentUserId,
  onCommentCreated,
}: CommentInputProps) => {
  const [content, setContent] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 파일 업로드 훅 사용
  const { selectedFile, handleFileSelect, handleRemoveFile } = useFileUpload();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim() && !selectedFile) return;

    setIsSubmitting(true);

    try {
      const attachmentIds: string[] = [];

      if (selectedFile) {
        const uploaded = await uploadAttachment(selectedFile, 'COMMENT', workspaceId);
        attachmentIds.push(uploaded.id);
      }

      await createComment({
        boardId: boardId,
        content: content.trim(),
        attachmentIds: attachmentIds,
      });

      setContent('');
      handleRemoveFile();
      if (fileInputRef.current) fileInputRef.current.value = ''; // input 초기화
      onCommentCreated();
    } catch (error) {
      console.error('댓글 작성 실패:', error);
      alert('댓글을 등록하지 못했습니다.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="mb-4 p-3 bg-white border border-gray-200 rounded-lg shadow-sm"
    >
      {/* 텍스트 입력 영역 */}
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="댓글을 입력하세요..."
        className="w-full text-sm text-gray-800 placeholder-gray-400 border-none focus:ring-0 resize-none p-1"
        rows={2} // 높이 줄임
        disabled={isSubmitting}
        style={{ outline: 'none' }}
      />

      <div className="mt-2 flex items-center justify-between border-t border-gray-100 pt-2">
        {/* 왼쪽: 파일 첨부 버튼 및 선택된 파일 표시 */}
        <div className="flex items-center gap-2 overflow-hidden">
          <input
            type="file"
            ref={fileInputRef}
            onChange={(e) => {
              // useFileUpload의 핸들러에 이벤트 전달
              // (FileUploader 컴포넌트를 안쓰므로 직접 연결)
              if (e.target.files && e.target.files.length > 0) {
                // useFileUpload 내부 로직에 맞춰 파일 객체만 넘기거나
                // hook이 event를 받는지 확인 필요.
                // 보통 hook이 event를 받는다면: handleFileSelect(e)
                handleFileSelect(e as any);
              }
            }}
            className="hidden"
            accept="image/*, .pdf, .doc, .docx, .xls, .xlsx, .ppt, .pptx, .txt"
          />

          <button
            type="button"
            onClick={() => fileInputRef.current?.click()}
            className="text-gray-400 hover:text-blue-500 transition p-1 rounded-full hover:bg-gray-100"
            title="파일 첨부"
          >
            <Paperclip size={18} />
          </button>

          {selectedFile && (
            <div className="flex items-center gap-1 px-2 py-1 bg-blue-50 text-blue-700 rounded-full text-xs max-w-[200px]">
              <span className="truncate max-w-[120px]">{selectedFile.name}</span>
              <button
                type="button"
                onClick={() => {
                  handleRemoveFile();
                  if (fileInputRef.current) fileInputRef.current.value = '';
                }}
                className="text-blue-400 hover:text-blue-600"
              >
                <X size={12} />
              </button>
            </div>
          )}
        </div>

        {/* 오른쪽: 전송 버튼 */}
        <button
          type="submit"
          disabled={isSubmitting || (!content.trim() && !selectedFile)}
          className="flex items-center gap-1 px-3 py-1.5 bg-blue-600 text-white text-xs font-bold rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {isSubmitting ? (
            '...'
          ) : (
            <>
              <Send size={12} /> 등록
            </>
          )}
        </button>
      </div>
    </form>
  );
};

// =============================================================================
// [Sub Component] 개별 댓글 아이템 (Compact Edit Mode)
// =============================================================================
interface CommentItemProps {
  comment: CommentResponse;
  nickname: string;
  profileUrl: string | null;
  workspaceId: string;
  currentUserId: string;
  onRefresh: () => void;
}

const CommentItem = ({
  comment,
  nickname,
  profileUrl,
  workspaceId,
  currentUserId,
  onRefresh,
}: CommentItemProps) => {
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(comment.content);
  const [isLoading, setIsLoading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const isMyComment = comment.userId === currentUserId;

  const existingAttachment =
    comment.attachments && comment.attachments.length > 0 ? comment.attachments[0] : null;

  const { selectedFile, handleFileSelect, handleRemoveFile, setInitialFile } = useFileUpload();

  useEffect(() => {
    if (isEditing && existingAttachment) {
      setInitialFile(existingAttachment.fileUrl, existingAttachment.fileName);
    }
  }, [isEditing, existingAttachment, setInitialFile]);

  const getUserColor = (name: string) => {
    const colors = [
      'bg-indigo-500',
      'bg-pink-500',
      'bg-green-500',
      'bg-purple-500',
      'bg-yellow-500',
    ];
    return colors[name.length % colors.length];
  };

  const handleUpdate = async () => {
    // 기존 파일 이름이 없고(삭제됨), 새 파일도 없고, 내용도 없으면 리턴
    const hasExisting = !!existingAttachment?.fileName; // hook의 previewUrl 등으로 체크 가능하지만 간단히
    // 여기서 previewUrl이나 selectedFile이 없으면 파일이 삭제된 것으로 간주해야 함.

    if (!editContent.trim() && !selectedFile && !existingAttachment) {
      // 로직 단순화
      alert('내용을 입력해주세요.');
      return;
    }

    setIsLoading(true);
    try {
      let attachmentIds: string[] = [];

      if (selectedFile) {
        const uploaded = await uploadAttachment(selectedFile, 'COMMENT', workspaceId);
        attachmentIds = [uploaded.id];
      } else if (existingAttachment) {
        // 파일 변경 없음 (기존 유지)
        // 기존 파일 삭제 로직을 구현하려면 useFileUpload에 'isDeleted' 같은 상태가 있거나
        // previewUrl이 null인지 체크해야 함. 여기서는 단순화하여 기존 파일 유지.
        attachmentIds = [existingAttachment.id];
      }

      await updateComment(comment.commentId, {
        content: editContent.trim(),
        attachmentIds: attachmentIds,
      });

      setIsEditing(false);
      onRefresh();
    } catch (error) {
      alert('수정 실패');
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!window.confirm('삭제하시겠습니까?')) return;
    setIsLoading(true);
    try {
      await deleteComment(comment.commentId);
      onRefresh();
    } catch (error) {
      alert('삭제 실패');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="p-3 bg-gray-50/50 border border-gray-100 rounded-lg group hover:bg-gray-100 transition-colors">
      <div className="flex items-start gap-3">
        {/* 아바타 */}
        <div className="w-8 h-8 rounded-full flex-shrink-0 overflow-hidden ring-1 ring-gray-200 bg-gray-200">
          {profileUrl ? (
            <img src={profileUrl} alt={nickname} className="w-full h-full object-cover" />
          ) : (
            <div
              className={`w-full h-full flex items-center justify-center text-white text-xs font-bold ${getUserColor(
                nickname,
              )}`}
            >
              {nickname?.[0] || '?'}
            </div>
          )}
        </div>

        {/* 내용 */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between mb-1">
            <div className="flex items-center gap-2">
              <span className="text-sm font-semibold text-gray-900">{nickname}</span>
              <span className="text-xs text-gray-400">
                {new Date(comment.createdAt).toLocaleDateString()}
              </span>
            </div>

            {isMyComment && !isEditing && (
              <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                  onClick={() => setIsEditing(true)}
                  className="p-1 text-gray-400 hover:text-blue-500 rounded"
                >
                  <Pencil size={12} />
                </button>
                <button
                  onClick={handleDelete}
                  className="p-1 text-gray-400 hover:text-red-500 rounded"
                >
                  <Trash2 size={12} />
                </button>
              </div>
            )}
          </div>

          {isEditing ? (
            <div className="mt-1">
              <textarea
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                className="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 resize-none"
                rows={2}
              />
              {/* 수정 모드 파일 버튼 (Compact) */}
              <div className="flex items-center justify-between mt-2">
                <div className="flex items-center gap-2">
                  <input
                    type="file"
                    ref={fileInputRef}
                    onChange={(e) => handleFileSelect(e as any)}
                    className="hidden"
                  />
                  <button
                    onClick={() => fileInputRef.current?.click()}
                    className="text-gray-500 hover:text-blue-600 p-1 bg-gray-200 rounded"
                  >
                    <Paperclip size={14} />
                  </button>
                  {(selectedFile || existingAttachment) && (
                    <span className="text-xs text-gray-600 truncate max-w-[150px]">
                      {selectedFile ? selectedFile.name : existingAttachment?.fileName}
                    </span>
                  )}
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => setIsEditing(false)}
                    className="text-xs px-2 py-1 text-gray-500 hover:bg-gray-200 rounded"
                  >
                    취소
                  </button>
                  <button
                    onClick={handleUpdate}
                    className="text-xs px-2 py-1 bg-blue-500 text-white rounded hover:bg-blue-600"
                  >
                    저장
                  </button>
                </div>
              </div>
            </div>
          ) : (
            <div>
              <p className="text-sm text-gray-800 whitespace-pre-wrap leading-relaxed">
                {comment.content}
              </p>
              {existingAttachment && (
                <div className="mt-2">
                  {existingAttachment.contentType.startsWith('image/') ? (
                    <a
                      href={existingAttachment.fileUrl}
                      target="_blank"
                      rel="noreferrer"
                      className="block max-w-[200px]"
                    >
                      <img
                        src={existingAttachment.fileUrl}
                        alt={existingAttachment.fileName}
                        className="rounded-md border border-gray-200 shadow-sm hover:opacity-90 transition"
                      />
                    </a>
                  ) : (
                    <a
                      href={existingAttachment.fileUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-xs text-blue-600 bg-blue-50 px-2 py-1 rounded hover:bg-blue-100 transition"
                    >
                      <Paperclip size={12} /> {existingAttachment.fileName}
                    </a>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

// =============================================================================
// [Main Component] 댓글 리스트
// =============================================================================
interface CommentListProps {
  boardId: string;
  workspaceId: string;
  comments: CommentResponse[];
  members: WorkspaceMemberResponse[];
  currentUserId: string;
  onRefresh: () => void;
}

const CommentList = ({
  boardId,
  workspaceId,
  comments,
  members,
  currentUserId,
  onRefresh,
}: CommentListProps) => {
  const { getNickname, getProfileUrl } = useUserLookup(members);

  return (
    <div className="flex flex-col h-full">
      {/* 댓글 목록 영역 - 💡 높이 확장 (max-h-80 -> flex-1 or max-h-[600px]) */}
      {/* 3개가 묻히지 않게 flex-1로 남은 공간을 다 쓰거나, 충분히 큰 max-h를 줍니다. */}
      <div className="flex-1 overflow-y-auto min-h-[85px] max-h-[500px] pr-2 custom-scrollbar space-y-1">
        {comments.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-10 text-gray-400">
            <p className="text-sm">아직 댓글이 없습니다.</p>
            <p className="text-xs">첫 번째 의견을 남겨보세요!</p>
          </div>
        ) : (
          comments.map((comment) => (
            <CommentItem
              key={comment.commentId}
              comment={comment}
              nickname={getNickname(comment.userId)}
              profileUrl={getProfileUrl(comment.userId)}
              workspaceId={workspaceId}
              currentUserId={currentUserId}
              onRefresh={onRefresh}
            />
          ))
        )}
      </div>

      {/* 댓글 작성 영역 */}
      <div className="flex-shrink-0 mt-1">
        <CommentInput
          boardId={boardId}
          workspaceId={workspaceId}
          currentUserId={currentUserId}
          onCommentCreated={onRefresh}
        />
      </div>
    </div>
  );
};

export default CommentList;
