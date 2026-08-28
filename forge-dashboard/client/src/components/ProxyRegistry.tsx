/**
 * T10 — Proxy registry UI.
 * Ce composant est un client mémoire seule du contrat référentiel proxy du
 * Core : aucune écriture SQLite directe, aucun credential (le champ
 * secret_ref pointe vers le vault du Core ; has_secret est un indicateur de
 * présence seulement). Chaque mutation porte un X-Request-ID et le Core
 * répond avec le X-Correlation-ID.
 */
import { FormEvent, useState } from "react";
import { AlertTriangle, LoaderCircle, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import {
  CoreProxy,
  CoreProxyClient,
  CoreProxyType,
  CoreWriteClient,
} from "@/lib/coreWrite";

const PORT_MIN = 1;
const PORT_MAX = 65535;

export type ProxyRegistryProps = {
  client: CoreWriteClient;
  proxies: CoreProxy[];
  onChange: (next: CoreProxy[]) => void;
  onAssigned?: (proxyId: string) => void;
  onUnassigned?: (proxyId: string) => void;
  /** Profil actuellement sélectionné dans le registre des profils, quand il est connu. */
  selectedProfileId?: string;
  /** Id du proxy déjà affecté au profil sélectionné (affichage uniquement). */
  assignedProxyId?: string;
  /**
   * État persistant du formulaire de création, fourni par le parent pour survivre
   * aux démontages/remontages du panneau (par exemple après l'expiration de la
   * session lecture seule du Core suivie d'un re-link du jeton d'administration).
   * Le contenu saisi ne quitte jamais la mémoire du navigateur ; le rechargement
   * de la page détruit le ref, comme le reste du contrôle local.
   */
  formRef: React.MutableRefObject<{
    name: string;
    type: CoreProxyType;
    host: string;
    port: string;
    region: string;
    secretRef: string;
  }>;
};

function humanizeError(error: unknown): string {
  if (!(error instanceof Error)) return "Erreur inattendue côté Core.";
  const message = error.message;
  if (message.startsWith("CORE_ERROR_")) {
    const code = message.slice("CORE_ERROR_".length);
    if (code === "PROXY_NAME_TAKEN") return "Ce proxy est encore affecté à un profil. Retirez l’affectation d’abord.";
    if (code === "INVALID_PROXY") return "Entrée refusée par le Core (type, hôte, port ou référence de secret non conforme).";
    if (code === "PROXY_LOCKED") return "Le proxy est verrouillé par une opération en cours. Réessayez.";
    if (code === "PROXY_NOT_FOUND") return "Ce proxy n’existe plus dans le référentiel du Core.";
    if (code === "PROFILE_NOT_FOUND") return "Ce profil n’existe pas dans le Core : l’affectation exige un profil réel du registre local.";
    if (code === "LOOPBACK_REQUIRED") return "Mutation hors loopback refusée : ouvrez le dashboard depuis la machine du Core.";
    if (code === "CORE_ADMIN_NOT_CONNECTED" || code === "CORE_ADMIN_UNAUTHORIZED") return "Le contrôle local a été retiré ; reliez le jeton d’administration.";
    if (code === "SUBMIT_TIMEOUT") return "Le Core ne répond pas sous 20 s : vérifiez qu’il est toujours démarré, puis réessayez.";
    return `Le Core a refusé l’opération (${code}).`;
  }
  if (message === "INVALID_PROXY_PORT") return "Le port doit être un entier entre 1 et 65535.";
  if (message === "INVALID_PROXY_HOST") return "L’hôte du proxy est requis.";
  if (message === "INVALID_PROXY_TYPE") return "Seuls les types http et socks5 sont acceptés.";
  if (message === "INVALID_PROXY_SECRET_REF") return "La référence de secret doit suivre la grammaire proxy.ref.* du vault.";
  if (message === "MISSING_NAME") return "Le nom du proxy est requis.";
  if (message === "CORE_NOT_LOOPBACK") return "Les écritures exigent une URL loopback du Core.";
  return "Connexion impossible : le Core local ne répond pas.";
}

function CoreProxyRow({
  proxy,
  client,
  selectedProfileId,
  isAssigned,
  onDelete,
  onAssigned,
  onUnassigned,
}: {
  proxy: CoreProxy;
  client: CoreWriteClient;
  selectedProfileId?: string;
  isAssigned: boolean;
  onDelete: () => void;
  onAssigned: () => void;
  onUnassigned: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const profileContext = selectedProfileId ?? "current-profile";
  const remove = async () => {
    setBusy(true);
    try {
      await client.proxies.deleteProxy(proxy.id);
      onDelete();
      toast.success("Proxy retiré du référentiel Core.");
    } catch (error) {
      toast.error(humanizeError(error));
    } finally {
      setBusy(false);
    }
  };
  return (
    <div className="proxy-row" data-testid={`proxy-row-${proxy.id}`}>
      <div className="proxy-row-main">
        <strong>{proxy.name}</strong>
        <span className="proxy-row-type">{proxy.type}</span>
        <span className="proxy-row-host">{proxy.host}:{proxy.port}</span>
        {proxy.region ? <span className="proxy-row-region">{proxy.region}</span> : null}
      </div>
      <div className="proxy-row-secrets">
        {proxy.has_secret ? (
          <span className="proxy-has-secret" title="Le vault du Core détient un secret rattaché à ce proxy. Sa valeur n’est jamais affichée.">
            Secret vault rattaché
          </span>
        ) : (
          <span className="proxy-no-secret">Aucun secret</span>
        )}
      </div>
      <div className="proxy-row-actions">
        {!isAssigned ? (
          <button
            data-testid={`proxy-assign-${proxy.id}`}
            disabled={busy || !selectedProfileId}
            title={selectedProfileId ? "Affecter ce proxy au profil sélectionné" : "Sélectionnez un profil Core d’abord"}
            type="button"
            onClick={async () => {
              setBusy(true);
              if (!selectedProfileId) {
                toast.error("Aucun profil Core n’est sélectionné : l’affectation exige un profil du Core.");
                return;
              }
              try {
                await client.proxies.assignProxy(proxy.id, selectedProfileId);
                onAssigned();
                toast.success("Proxy affecté au profil côté Core.");
              } catch (error) {
                toast.error(humanizeError(error));
              } finally {
                setBusy(false);
              }
            }}
          >
            Affecter
          </button>
        ) : (
          <button
            data-testid={`proxy-unassign-${proxy.id}`}
            disabled={busy}
            type="button"
            onClick={async () => {
              setBusy(true);
              if (!selectedProfileId) {
                toast.error("Aucun profil Core n’est sélectionné : le retrait exige un profil du Core.");
                return;
              }
              try {
                await client.proxies.unassignProxy(proxy.id, selectedProfileId);
                onUnassigned();
                toast.success("Affectation retirée côté Core.");
              } catch (error) {
                toast.error(humanizeError(error));
              } finally {
                setBusy(false);
              }
            }}
          >
            Retirer affectation
          </button>
        )}
        <button
          data-testid={`proxy-delete-${proxy.id}`}
          disabled={busy}
          type="button"
          onClick={remove}
        >
          {busy ? <LoaderCircle className="spin" size={12} /> : <Trash2 size={12} />}
          Retirer
        </button>
      </div>
    </div>
  );
}

export function ProxyRegistry({ client, proxies, onChange, onAssigned, onUnassigned, selectedProfileId, assignedProxyId, formRef }: ProxyRegistryProps) {
  const [creating, setCreating] = useState(false);

  // Double source pilotée par un state local :
  // - le state garantit un re-render à CHAQUE saisie (un setter à valeur
  //   identique bailouterait), donc le bouton réagit immédiatement ;
  // - le ref partagé fourni par le parent persiste le contenu entre remounts
  //   du panneau (expiration de la session lecture seule suivie d'un re-link),
  //   et reste la valeur lue au submit.
  const [form, setForm] = useState(() => formRef.current);
  const keepForm = (next: typeof formRef.current) => {
    formRef.current = next;
    setForm(next);
  };
  // Les mutations du panneau peuvent provoquer un rerender entre deux saisies
  // Playwright ou utilisateur. Toujours dériver la prochaine valeur de la ref
  // (et non de la fermeture `form`) empêche qu’une saisie rapide réintroduise
  // un champ vide d’un rendu antérieur et bloque le bouton de création.
  const updateForm = (patch: Partial<typeof form>) => {
    keepForm({ ...formRef.current, ...patch });
  };

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (creating) return;
    const submittedForm = formRef.current;
    const numericPort = Number(submittedForm.port);
    setCreating(true);
    try {
      // Fail-safe : la soumission est bornée dans le temps. Si le Core ne répond
      // pas sous 20 s (surcharge, runtime en cours de qualification, arrêt du
      // Core), le formulaire redevient utilisable et une erreur explicite est
      // affichée au lieu d'un bouton gelé sans aucun retour.
      const result = await Promise.race([
        client.proxies.createProxy({
          name: submittedForm.name.trim(),
          type: submittedForm.type,
          host: submittedForm.host.trim(),
          port: Number.isFinite(numericPort) && numericPort >= PORT_MIN && numericPort <= PORT_MAX ? numericPort : 0,
          region: submittedForm.region.trim() || undefined,
          secret_ref: submittedForm.secretRef.trim() || undefined,
        }),
        new Promise<never>((_resolve, reject) =>
          window.setTimeout(() => reject(new Error("CORE_ERROR_SUBMIT_TIMEOUT")), 20_000),
        ),
      ]);
      onChange([...proxies, result.data]);
      keepForm({ name: "", type: "http" as CoreProxyType, host: "", port: "", region: "", secretRef: "" });
      toast.success("Proxy ajouté au référentiel Core.", { description: "correlation_id: " + (result.correlationId ?? "—") });
    } catch (error) {
      toast.error(humanizeError(error));
    } finally {
      setCreating(false);
    }
  };

  return (
    <section className="proxy-registry" data-testid="proxy-registry">
      <div className="proxy-registry-title">
        <Plus size={13} />
        Référentiel proxy
      </div>
      <p className="proxy-registry-note">
        Le référentiel est la propriété du Core. Le dashboard ne saisit que des
        métadonnées : aucune valeur de credential n’est envoyée, affichée ou
        conservée côté navigateur. Les identifiants passent par le vault du
        Core (proxy.ref.*).
      </p>
      {proxies.length === 0 ? (
        <p className="proxy-empty" data-testid="proxy-empty">
          <AlertTriangle size={12} />
          Aucun proxy enregistré. Créez-en un pour pouvoir l’affecter aux profils.
        </p>
      ) : (
        <div className="proxy-list">
          {proxies.map(proxy => (
            <CoreProxyRow
              key={proxy.id}
              proxy={proxy}
              client={client}
              selectedProfileId={selectedProfileId}
              isAssigned={assignedProxyId === proxy.id}
              onDelete={() => onChange(proxies.filter(p => p.id !== proxy.id))}
              onAssigned={() => {
                onChange(proxies);
                onAssigned?.(proxy.id);
              }}
              onUnassigned={() => {
                onChange(proxies);
                onUnassigned?.(proxy.id);
              }}
            />
          ))}
        </div>
      )}
      <form className="proxy-create-form" onSubmit={submit}>
        <input
          autoComplete="off"
          data-testid="proxy-name"
          maxLength={64}
          onChange={event =>
            updateForm({ name: event.target.value })
          }
          placeholder="Nom du proxy"
          spellCheck={false}
          value={form.name}
        />
        <select data-testid="proxy-type" value={form.type} onChange={event =>
          updateForm({ type: event.target.value as CoreProxyType })
        }>
          <option value="http">http</option>
          <option value="socks5">socks5</option>
        </select>
        <input
          autoComplete="off"
          data-testid="proxy-host"
          maxLength={128}
          onChange={event =>
            updateForm({ host: event.target.value })
          }
          placeholder="Hôte (ex. proxy.local)"
          spellCheck={false}
          value={form.host}
        />
        <input
          autoComplete="off"
          data-testid="proxy-port"
          inputMode="numeric"
          maxLength={5}
          onChange={event =>
            updateForm({ port: event.target.value.replace(/\D/g, "") })
          }
          placeholder="Port (1–65535)"
          value={form.port}
        />
        <input
          autoComplete="off"
          data-testid="proxy-region"
          maxLength={32}
          onChange={event =>
            updateForm({ region: event.target.value })
          }
          placeholder="Région (facultatif)"
          spellCheck={false}
          value={form.region}
        />
        <input
          autoComplete="off"
          data-testid="proxy-secret-ref"
          maxLength={160}
          onChange={event =>
            updateForm({ secretRef: event.target.value })
          }
          placeholder="Référence vault proxy.ref.* (facultatif)"
          spellCheck={false}
          value={form.secretRef}
        />
        <button data-testid="proxy-create" disabled={creating || !form.name.trim() || !form.host.trim() || !form.port.trim()} type="submit">
          {creating ? <LoaderCircle className="spin" size={12} /> : <Plus size={12} />}
          Ajouter au Core
        </button>
      </form>
    </section>
  );
}
