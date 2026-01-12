import { Header } from './Header';

/**
 * Props for Layout component
 */
interface LayoutProps {
  children: React.ReactNode;
}

/**
 * Layout component that wraps protected pages with header
 */
export const Layout = ({ children }: LayoutProps) => {
  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="container mx-auto px-4 py-6">
        {children}
      </main>
    </div>
  );
};
