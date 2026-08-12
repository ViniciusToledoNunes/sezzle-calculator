import { type FormEvent, useMemo, useRef, useState } from "react";
import { ApiError, calculate } from "./api/calculator";
import {
  OPERATIONS,
  formatExpression,
  formatNumber,
  getOperation,
  type Operation,
} from "./domain/operations";
import "./App.css";

type FieldErrors = { first?: string; second?: string };
type HistoryItem = { id: number; expression: string; result: number };

function parseNumber(value: string, label: string): { value?: number; error?: string } {
  if (value.trim() === "") {
    return { error: `Enter the ${label} number.` };
  }
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) {
    return { error: `The ${label} value must be a valid number.` };
  }
  return { value: parsed };
}

export default function App() {
  const [operationId, setOperationId] = useState<Operation>("add");
  const [firstValue, setFirstValue] = useState("");
  const [secondValue, setSecondValue] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [result, setResult] = useState<number | null>(null);
  const [requestError, setRequestError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [history, setHistory] = useState<HistoryItem[]>([]);
  const nextHistoryId = useRef(0);

  const operation = useMemo(() => getOperation(operationId), [operationId]);

  function chooseOperation(nextOperation: Operation) {
    setOperationId(nextOperation);
    setFieldErrors({});
    setRequestError("");
    setResult(null);
  }

  function clearCalculator() {
    setFirstValue("");
    setSecondValue("");
    setFieldErrors({});
    setRequestError("");
    setResult(null);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isLoading) return;
    setRequestError("");

    const first = parseNumber(firstValue, "first");
    const second = operation.arity === 2 ? parseNumber(secondValue, "second") : {};
    const errors: FieldErrors = { first: first.error, second: second.error };
    setFieldErrors(errors);
    if (errors.first || errors.second || first.value === undefined) {
      return;
    }

    const operands = operation.arity === 2 ? [first.value, second.value as number] : [first.value];
    setIsLoading(true);
    try {
      const nextResult = await calculate(operation.id, operands);
      setResult(nextResult);
      setHistory((current) => [
        {
          id: ++nextHistoryId.current,
          expression: formatExpression(operation, operands),
          result: nextResult,
        },
        ...current,
      ].slice(0, 5));
    } catch (error) {
      setResult(null);
      setRequestError(
        error instanceof ApiError || error instanceof Error
          ? error.message
          : "Something went wrong. Please try again.",
      );
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="app-shell">
      <header className="site-header">
        <a className="brand" href="#calculator" aria-label="Sezzle Calculator home">
          <span className="brand-mark" aria-hidden="true"><i /><i /><i /></span>
          <span>Sezzle <strong>Calculator</strong></span>
        </a>
        <span className="api-status"><i aria-hidden="true" /> Powered by Go API</span>
      </header>

      <main className="page" id="calculator">
        <section className="intro" aria-labelledby="page-title">
          <p className="eyebrow">Simple math. Reliable results.</p>
          <h1 id="page-title">Calculate with confidence.</h1>
          <p>Choose an operation, enter your values, and let our tested backend do the math.</p>
        </section>

        <div className="workspace">
          <section className="calculator-card" aria-labelledby="calculator-title">
            <div className="card-heading">
              <div>
                <p className="card-kicker">New calculation</p>
                <h2 id="calculator-title">What should we solve?</h2>
              </div>
              <button className="text-button" type="button" disabled={isLoading} onClick={clearCalculator}>Clear</button>
            </div>

            <form onSubmit={handleSubmit} noValidate>
              <fieldset className="operations">
                <legend>Operation</legend>
                <div className="operation-grid">
                  {OPERATIONS.map((candidate) => (
                    <button
                      className="operation-button"
                      type="button"
                      key={candidate.id}
                      disabled={isLoading}
                      aria-pressed={candidate.id === operationId}
                      aria-label={candidate.label}
                      title={candidate.description}
                      onClick={() => chooseOperation(candidate.id)}
                    >
                      <span aria-hidden="true">{candidate.symbol}</span>
                      <small>{candidate.label}</small>
                    </button>
                  ))}
                </div>
              </fieldset>

              <div className={`input-grid ${operation.arity === 1 ? "input-grid--single" : ""}`}>
                <label className="number-field">
                  <span>{operation.arity === 1 ? "Number" : "First number"}</span>
                  <input
                    autoFocus
                    type="text"
                    inputMode="decimal"
                    autoComplete="off"
                    disabled={isLoading}
                    value={firstValue}
                    aria-invalid={Boolean(fieldErrors.first)}
                    aria-describedby={fieldErrors.first ? "first-error" : undefined}
                    placeholder="0"
                    onChange={(event) => {
                      setFirstValue(event.target.value);
                      if (fieldErrors.first) setFieldErrors((errors) => ({ ...errors, first: undefined }));
                    }}
                  />
                  {fieldErrors.first && <small className="field-error" id="first-error" role="alert">{fieldErrors.first}</small>}
                </label>

                {operation.arity === 2 && (
                  <label className="number-field">
                    <span>Second number</span>
                    <input
                      type="text"
                      inputMode="decimal"
                      autoComplete="off"
                      disabled={isLoading}
                      value={secondValue}
                      aria-invalid={Boolean(fieldErrors.second)}
                      aria-describedby={fieldErrors.second ? "second-error" : undefined}
                      placeholder="0"
                      onChange={(event) => {
                        setSecondValue(event.target.value);
                        if (fieldErrors.second) setFieldErrors((errors) => ({ ...errors, second: undefined }));
                      }}
                    />
                    {fieldErrors.second && <small className="field-error" id="second-error" role="alert">{fieldErrors.second}</small>}
                  </label>
                )}
              </div>

              <button className="calculate-button" type="submit" disabled={isLoading}>
                {isLoading ? <><span className="spinner" aria-hidden="true" /> Calculating…</> : "Calculate result"}
              </button>
            </form>

            <div className={`result-panel ${requestError ? "result-panel--error" : ""}`} aria-live="polite">
              <span className="result-label">Result</span>
              {requestError ? (
                <p className="request-error" role="alert">{requestError}</p>
              ) : result === null ? (
                <p className="result-placeholder">Your answer will appear here</p>
              ) : (
                <output className="result-value" aria-label={`Result: ${formatNumber(result)}`}>{formatNumber(result)}</output>
              )}
            </div>
          </section>

          <aside className="history-card" aria-labelledby="history-title">
            <div className="history-heading">
              <div>
                <p className="card-kicker">This session</p>
                <h2 id="history-title">Recent results</h2>
              </div>
              {history.length > 0 && (
                <button className="text-button" type="button" onClick={() => setHistory([])}>Clear</button>
              )}
            </div>

            {history.length === 0 ? (
              <div className="empty-history">
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M12 8v4l2.5 1.5M21 12a9 9 0 1 1-2.64-6.36M21 4v5h-5" />
                </svg>
                <p>No calculations yet</p>
                <span>Your five most recent results will stay here.</span>
              </div>
            ) : (
              <ol className="history-list">
                {history.map((item) => (
                  <li key={item.id}>
                    <span>{item.expression}</span>
                    <strong>{formatNumber(item.result)}</strong>
                  </li>
                ))}
              </ol>
            )}

            <div className="privacy-note">
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3 5 6v5c0 4.6 2.9 8.9 7 10 4.1-1.1 7-5.4 7-10V6l-7-3Z" /><path d="m9 12 2 2 4-4" /></svg>
              <p><strong>Private by design</strong><span>History stays in this browser tab and is never stored.</span></p>
            </div>
          </aside>
        </div>
      </main>

      <footer>Built with React, TypeScript, and Go.</footer>
    </div>
  );
}
