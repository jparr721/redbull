"use client";

import { useState } from "react";
import { DataTable } from "@/components/data-table";
import { getResponses, ResponseSchema } from "@/queries/responses.query";
import { useQuery, useMutation } from "@tanstack/react-query";
import { Input } from "@/components/ui/input";
import axios from "axios";
import { createColumnsFromSchema } from '../../lib/table-utils';

export function Responses() {
  const [command, setCommand] = useState("");

  const { data: responses, isLoading } = useQuery({
    queryKey: ["responses"],
    queryFn: getResponses,
    refetchInterval: 1000,
  });

  const mutation = useMutation({
    mutationFn: (cmd: string) => axios.post("http://localhost:8000/command", btoa(cmd)),
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!command.trim()) return;
    mutation.mutate(command);
    setCommand("");
  };

  if (isLoading) return <h1>Loading</h1>;

  return (
    <div>
      <form onSubmit={handleSubmit} className="mb-4">
        <Input
          value={command}
          onChange={(e) => setCommand(e.target.value)}
          placeholder="Enter command..."
          className="w-full"
        />
      </form>
      {!(!responses || responses.length === 0) && <DataTable data={responses} columns={createColumnsFromSchema(ResponseSchema)} />}
    </div>
  );
}
