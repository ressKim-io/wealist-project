import React, { useState } from 'react';
import { Search, Settings, ChevronDown } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';

interface FilterBarProps {
  onSearchChange: (search: string) => void;
  onViewChange: (view: 'stage' | 'role') => void;
  onFilterChange: (filter: string) => void;
  onManageClick: () => void;
  currentView: 'stage' | 'role';
}

export const FilterBar: React.FC<FilterBarProps> = ({
  onSearchChange,
  onViewChange,
  onFilterChange,
  onManageClick,
  currentView,
}) => {
  const { theme } = useTheme();
  const [searchValue, setSearchValue] = useState('');
  const [selectedFilter, setSelectedFilter] = useState('all');
  const [showViewDropdown, setShowViewDropdown] = useState(false);
  const [showFilterDropdown, setShowFilterDropdown] = useState(false);

  const handleSearchChange = (value: string) => {
    setSearchValue(value);
    onSearchChange(value);
  };

  const handleViewChange = (view: 'stage' | 'role') => {
    onViewChange(view);
    setShowViewDropdown(false);
  };

  const handleFilterChange = (filter: string) => {
    setSelectedFilter(filter);
    onFilterChange(filter);
    setShowFilterDropdown(false);
  };

  const viewOptions = [
    { value: 'stage', label: 'Stage 기준' },
    { value: 'role', label: 'Role 기준' },
  ];

  const filterOptions = [
    { value: 'all', label: '전체' },
    { value: 'my', label: '내가 담당한 것만' },
    { value: 'high', label: '중요도 높음' },
    { value: 'urgent', label: '긴급' },
    { value: 'hideCompleted', label: '완료된 것 숨기기' },
  ];

  return (
    <div
      className={`flex items-center gap-3 p-4 ${theme.colors.card} border-b ${theme.colors.border}`}
    >
      {/* Search */}
      <div className="relative flex-1 max-w-md">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
        <input
          type="text"
          value={searchValue}
          onChange={(e) => handleSearchChange(e.target.value)}
          placeholder="보드 검색..."
          className={`w-full pl-10 pr-4 py-2 border ${theme.colors.border} rounded-md ${theme.colors.card} focus:outline-none focus:ring-2 focus:ring-blue-500`}
        />
      </div>

      {/* View Selector */}
      <div className="relative">
        <button
          onClick={() => setShowViewDropdown(!showViewDropdown)}
          className={`flex items-center gap-2 px-4 py-2 border ${theme.colors.border} rounded-md ${theme.colors.card} hover:bg-gray-50 transition-colors`}
        >
          <span className="text-sm font-medium">
            뷰: {viewOptions.find((v) => v.value === currentView)?.label}
          </span>
          <ChevronDown className="w-4 h-4" />
        </button>
        {showViewDropdown && (
          <div
            className={`absolute top-full mt-2 right-0 w-48 ${theme.colors.card} border ${theme.colors.border} rounded-md shadow-lg z-10`}
          >
            {viewOptions.map((option) => (
              <button
                key={option.value}
                onClick={() => handleViewChange(option.value as 'stage' | 'role')}
                className={`w-full text-left px-4 py-2 hover:bg-gray-50 transition-colors ${
                  currentView === option.value ? 'bg-blue-50 text-blue-600 font-medium' : ''
                }`}
              >
                {option.label}
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Filter Dropdown */}
      <div className="relative">
        <button
          onClick={() => setShowFilterDropdown(!showFilterDropdown)}
          className={`flex items-center gap-2 px-4 py-2 border ${theme.colors.border} rounded-md ${theme.colors.card} hover:bg-gray-50 transition-colors`}
        >
          <span className="text-sm font-medium">
            필터: {filterOptions.find((f) => f.value === selectedFilter)?.label}
          </span>
          <ChevronDown className="w-4 h-4" />
        </button>
        {showFilterDropdown && (
          <div
            className={`absolute top-full mt-2 right-0 w-56 ${theme.colors.card} border ${theme.colors.border} rounded-md shadow-lg z-10`}
          >
            {filterOptions.map((option) => (
              <button
                key={option.value}
                onClick={() => handleFilterChange(option.value)}
                className={`w-full text-left px-4 py-2 hover:bg-gray-50 transition-colors ${
                  selectedFilter === option.value ? 'bg-blue-50 text-blue-600 font-medium' : ''
                }`}
              >
                {option.label}
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Manage Button */}
      <button
        onClick={onManageClick}
        className="flex items-center gap-2 px-4 py-2 bg-gray-100 hover:bg-gray-200 rounded-md transition-colors"
        title="커스텀 필드 관리"
      >
        <Settings className="w-4 h-4" />
        <span className="text-sm font-medium">관리</span>
      </button>
    </div>
  );
};
