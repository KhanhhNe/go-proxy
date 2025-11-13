import { AppSidebar } from "./components/app-sidebar";
import { SidebarProvider } from "./components/ui/sidebar";
import { PageIndex } from "./pages";
import { PAGES, usePageStore } from "./state";

function App() {
  const page = usePageStore((state) => state.page);

  return (
    <SidebarProvider>
      <AppSidebar />
      <main id="App" className="m-5">
        {page === PAGES.index && <PageIndex />}
      </main>
    </SidebarProvider>
  );
}

export default App;
