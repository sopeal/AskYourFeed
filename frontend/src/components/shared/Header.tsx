import { Link, useNavigate, useLocation } from 'react-router-dom';
import { LogOut } from 'lucide-react';
import { SyncStatusIndicator } from './SyncStatusIndicator';
import { Avatar, AvatarFallback } from '../ui/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu';

/**
 * Get user initials from display name or username
 */
const getUserInitials = (displayName: string, username: string): string => {
  if (displayName && displayName.trim()) {
    const parts = displayName.trim().split(' ');
    if (parts.length >= 2) {
      return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
    }
    return displayName.substring(0, 2).toUpperCase();
  }
  return username.substring(0, 2).toUpperCase();
};

/**
 * Header component with navigation, sync status, and user menu
 */
export const Header = () => {
  const navigate = useNavigate();
  const location = useLocation();

  // Get user data from localStorage
  const getUserData = () => {
    try {
      const userStr = localStorage.getItem('user');
      if (userStr) {
        return JSON.parse(userStr);
      }
    } catch (error) {
      console.error('Error parsing user data:', error);
    }
    return null;
  };

  const user = getUserData();

  // Handle logout
  const handleLogout = () => {
    localStorage.removeItem('session_token');
    localStorage.removeItem('user');
    navigate('/login');
  };

  // Check if current path is active
  const isActive = (path: string) => location.pathname === path;

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-16 items-center justify-between px-4">
        {/* Logo and Navigation */}
        <div className="flex items-center gap-6">
          <Link to="/" className="flex items-center space-x-2">
            <span className="text-xl font-bold">AskYourFeed</span>
          </Link>
          
          <nav className="flex items-center gap-4">
            <Link
              to="/"
              className={`text-sm font-medium transition-colors hover:text-primary ${
                isActive('/') 
                  ? 'text-foreground' 
                  : 'text-muted-foreground'
              }`}
            >
              Panel Główny
            </Link>
            <Link
              to="/history"
              className={`text-sm font-medium transition-colors hover:text-primary ${
                isActive('/history') 
                  ? 'text-foreground' 
                  : 'text-muted-foreground'
              }`}
            >
              Historia
            </Link>
          </nav>
        </div>

        {/* Sync Status and User Menu */}
        <div className="flex items-center gap-4">
          {/* Sync Status Indicator */}
          <SyncStatusIndicator />

          {/* User Menu */}
          {user && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button className="flex items-center gap-2 rounded-full focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2">
                  <Avatar className="h-8 w-8 cursor-pointer">
                    <AvatarFallback className="bg-primary text-primary-foreground">
                      {getUserInitials(user.x_display_name || '', user.x_username || '')}
                    </AvatarFallback>
                  </Avatar>
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuLabel>
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none">
                      {user.x_display_name || user.x_username}
                    </p>
                    <p className="text-xs leading-none text-muted-foreground">
                      @{user.x_username}
                    </p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleLogout} className="cursor-pointer">
                  <LogOut className="mr-2 h-4 w-4" />
                  <span>Wyloguj</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </div>
    </header>
  );
};
