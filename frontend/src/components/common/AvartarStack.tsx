// =============================================================================
// AvatarStack (변경 없음)

import { WorkspaceMember } from '../../types/user';

// =============================================================================
interface AvatarStackProps {
  members: WorkspaceMember[];
}

export const AvatarStack: React.FC<AvatarStackProps> = ({ members }) => {
  const displayCount = 3;
  const displayMembers = members.slice(0, displayCount);
  const remainingCount = members.length - displayCount;

  const getColorByIndex = (index: number) => {
    const colors = [
      'bg-indigo-500',
      'bg-pink-500',
      'bg-green-500',
      'bg-purple-500',
      'bg-yellow-500',
    ];
    return colors[index % colors.length];
  };

  return (
    <div className="flex -space-x-1.5 p-1 pr-0 overflow-hidden">
      {displayMembers.map((member, index) => (
        <div
          key={member.userId}
          className="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white overflow-hidden"
          style={{ zIndex: members.length - index }}
          title={`${member.userName} (${member.roleName})`}
        >
          {member.profileImageUrl ? (
            <img
              src={member.profileImageUrl}
              alt={member.userName}
              className="w-full h-full object-cover"
            />
          ) : (
            <div
              className={`w-full h-full flex items-center justify-center text-white ${getColorByIndex(
                index,
              )}`}
            >
              {member.userName[0]}
            </div>
          )}
        </div>
      ))}
      {remainingCount > 0 && (
        <div
          className="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white bg-gray-400 text-white"
          style={{ zIndex: 0 }}
        >
          +{remainingCount}
        </div>
      )}
    </div>
  );
};

interface AssigneeAvatarStackProps {
  assignees: string | string[];
}

export const AssigneeAvatarStack: React.FC<AssigneeAvatarStackProps> = ({ assignees }) => {
  const assigneeList = Array.isArray(assignees)
    ? assignees
    : (assignees as string)
        .split(',')
        .map((name) => name.trim())
        .filter((name) => name.length > 0);

  const initials = assigneeList.map((name) => name[0]).filter((i) => i);
  const displayCount = 3;

  if (initials.length === 0) {
    return (
      <div
        className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-gray-200 bg-gray-200 text-gray-700`}
      >
        ?
      </div>
    );
  }

  return (
    <div className="flex -space-x-1 p-1 pr-0 overflow-hidden">
      {initials.slice(0, displayCount).map((initial, index) => (
        <div
          key={index}
          className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white text-white ${
            index === 0 ? 'bg-indigo-500' : index === 1 ? 'bg-pink-500' : 'bg-green-500'
          }`}
          style={{ zIndex: initials.length - index }}
          title={assigneeList[index]}
        >
          {initial}
        </div>
      ))}
      {initials.length > displayCount && (
        <div
          className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white bg-gray-400 text-white`}
          style={{ zIndex: 0 }}
          title={`${initials.length - displayCount}명 외`}
        >
          +{initials.length - displayCount}
        </div>
      )}
    </div>
  );
};
