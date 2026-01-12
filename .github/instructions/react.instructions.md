---
applyTo: '**/*.ts,**/*.tsx'
---
The language for use is Vietnamese, always use Vietnamese for UI components, stuff that user see, except some specialized keywords that's easily understood (file, proxy, port, host, etc). All variables and code are still in English.

For UI components, prioritizes using components from shadcn/ui library (imported from @/components/ui/...) for consistency in design and functionality. Avoid using standard HTML elements when a shadcn/ui equivalent exists.

Avoid using useEffect if possible, prefer direct calculations, handlers, better to replace existing on<event> with new handler instead of useEffect to update