import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import Layout from "./components/Layout";
import ProductList from "./pages/ProductList";
import ProductDetail from "./pages/ProductDetail";
import BranchDetail from "./pages/BranchDetail";
import ReleaseDetail from "./pages/ReleaseDetail";

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, refetchOnWindowFocus: false } },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<Navigate to="/products" replace />} />
            <Route path="/products" element={<ProductList />} />
            <Route path="/products/:productId" element={<ProductDetail />} />
            <Route path="/branches/:branchId" element={<BranchDetail />} />
            <Route path="/releases/:releaseId" element={<ReleaseDetail />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
