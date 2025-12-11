"use client";

import { useState } from "react";
import { getResponses } from "@/queries/responses.query";
import { useQuery, useMutation } from "@tanstack/react-query";
import { Input } from "@/components/ui/input";
import axios from "axios";
import { StickToBottom } from "use-stick-to-bottom";
import {
  Tool,
  ToolContent,
  ToolHeader,
  ToolOutput,
} from "@/components/ai-elements/tool";

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
    <div className="flex flex-col h-full">
      <form onSubmit={handleSubmit} className="mb-4">
        <Input
          value={command}
          onChange={(e) => setCommand(e.target.value)}
          placeholder="Enter command..."
          className="w-full"
        />
      </form>
      <StickToBottom className="relative flex-1 overflow-y-hidden">
        <StickToBottom.Content className="flex flex-col gap-4 p-4">
          {responses?.map((response, i) => (
            <Tool key={response.id} defaultOpen={i == responses.length - 1}>
              <ToolHeader
                title={response.command}
                type={`tool-${response.command}`}
                state={response.stderr.trim().length > 0 ? "output-error" : "output-available"}
              />
              <ToolContent>
                <ToolOutput
                  output={response.stdout}
                  errorText={response.stderr.trim().length > 0 ? response.stderr : undefined}
                />
              </ToolContent>
            </Tool>
          ))}
        </StickToBottom.Content>
      </StickToBottom>
    </div>
  );
}
