import { GetManager } from "@bindings/go-proxy/myservice";
import { useEffect } from "react";
import { AppSidebar } from "./components/app-sidebar";
import { SidebarProvider } from "./components/ui/sidebar";
import { PageIndex } from "./pages";
import { PageServers } from "./pages/servers";
import { PAGES, useManagerStore, usePageStore } from "./state";

function App() {
  const page = usePageStore((state) => state.page);
  const setManager = useManagerStore((state) => state.setManager);

  useEffect(() => {
    function fetchManager() {
      GetManager().then((m) => (m ? setManager(m) : null));

      setTimeout(fetchManager, 1000);
    }

    fetchManager();
  }, []);

  return (
    <SidebarProvider>
      <AppSidebar />
      <main id="App" className="m-5 w-full">
        {page === PAGES.index && <PageIndex />}
        {page === PAGES.servers && <PageServers />}
      </main>
    </SidebarProvider>
  );
}

export default App;
