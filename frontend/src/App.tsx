import "./App.css";
import { AppSidebar } from "./components/app-sidebar";
import { SidebarProvider } from "./components/ui/sidebar";

function App() {
  return (
    <SidebarProvider>
      <AppSidebar />
      <main id="App"></main>
    </SidebarProvider>
  );
}

export default App;
