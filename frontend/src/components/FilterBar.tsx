import React, { useState } from 'react';
import { Search, ChevronDown, Eye, Table, LayoutGrid } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';

interface FilterBarProps {
  onSearchChange: (search: string) => void;
  onViewChange: (view: 'stage' | 'role') => void;
  onFilterChange: (filter: string) => void;
  onManageClick: () => void;
  currentView: 'stage' | 'role';
  onLayoutChange?: (layout: 'table' | 'board') => void;
  onShowCompletedChange?: (show: boolean) => void;
  currentLayout?: 'table' | 'board';
  showCompleted?: boolean;
}

export const FilterBar: React.FC<FilterBarProps> = ({
  onSearchChange,
  onViewChange,
  onFilterChange,
  onManageClick,
  currentView,
  onLayoutChange,
  onShowCompletedChange,
  currentLayout = 'board',
  showCompleted = false,
}) => {
  const { theme } = useTheme();
  const [searchValue, setSearchValue] = useState('');
  const [selectedFilter, setSelectedFilter] = useState('all');
  const [showViewDropdown, setShowViewDropdown] = useState(false);
  const [showFilterDropdown, setShowFilterDropdown] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);

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
    { value: 'custom', label: '커스텀 필터' },
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

      {/* View Modal Button */}
      <div className="relative">
        <button
          onClick={() => {
            setShowViewModal(!showViewModal);
            setShowFilterDropdown(false);
          }}
          className={`flex items-center gap-2 px-4 py-2 border ${theme.colors.border} rounded-md ${theme.colors.card} hover:bg-gray-50 transition-colors`}
        >
          <Eye className="w-4 h-4" />
          <span className="text-sm font-medium">보기</span>
          <ChevronDown className={`w-4 h-4 transition-transform ${showViewModal ? 'rotate-180' : ''}`} />
        </button>
        {showViewModal && (
          <div
            className={`absolute top-full mt-2 right-0 w-64 ${theme.colors.card} border ${theme.colors.border} rounded-lg shadow-lg z-10 p-4`}
          >
            {/* Layout Selection */}
            <div className="mb-4">
              <h4 className="text-xs font-semibold text-gray-500 mb-2">레이아웃</h4>
              <div className="flex gap-2">
                <button
                  onClick={() => {
                    onLayoutChange?.('table');
                  }}
                  className={`flex-1 flex flex-col items-center gap-2 px-3 py-3 rounded-md border-2 transition-all ${
                    currentLayout === 'table'
                      ? 'border-blue-500 bg-blue-50 text-blue-600'
                      : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                  }`}
                >
                  <Table className={`w-6 h-6 ${currentLayout === 'table' ? 'text-blue-600' : 'text-gray-600'}`} />
                  <span className="text-sm font-medium">표</span>
                </button>
                <button
                  onClick={() => {
                    onLayoutChange?.('board');
                  }}
                  className={`flex-1 flex flex-col items-center gap-2 px-3 py-3 rounded-md border-2 transition-all ${
                    currentLayout === 'board'
                      ? 'border-blue-500 bg-blue-50 text-blue-600'
                      : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                  }`}
                >
                  <LayoutGrid className={`w-6 h-6 ${currentLayout === 'board' ? 'text-blue-600' : 'text-gray-600'}`} />
                  <span className="text-sm font-medium">보드</span>
                </button>
              </div>
            </div>

            {/* Divider */}
            <div className="border-t border-gray-200 my-3"></div>

            {/* Show Completed Toggle */}
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-700">완료된 항목 보기</span>
              <button
                onClick={() => {
                  onShowCompletedChange?.(!showCompleted);
                }}
                className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                  showCompleted ? 'bg-blue-500' : 'bg-gray-300'
                }`}
              >
                <span
                  className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    showCompleted ? 'translate-x-6' : 'translate-x-1'
                  }`}
                />
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Filter Dropdown */}
      <div className="relative">
        <button
          onClick={() => {
            setShowFilterDropdown(!showFilterDropdown);
            setShowViewModal(false);
          }}
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
    </div>
  );
};
