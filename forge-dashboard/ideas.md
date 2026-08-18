# Direction de design — ForgeLocal Dashboard

## Trois approches explorées

### 1. Atelier de contrôle local

Un environnement de bureau calme, dense et tangible, inspiré des établis techniques et des consoles d’exploitation. Il transmet la maîtrise locale sans le vocabulaire visuel agressif de la cybersécurité.

**Probabilité :** 0,07

### 2. Registre opérationnel

Une interface éditoriale lumineuse, proche d’un registre contemporain : papier minéral, encre sombre, annotations et hiérarchie documentaire. Elle mettrait l’accent sur l’auditabilité et la lisibilité des décisions.

**Probabilité :** 0,04

### 3. Salle de routage nocturne

Un tableau de bord sombre, à contrastes contrôlés, inspiré des panneaux industriels et du monitoring de réseau. Les états deviennent des signaux rares et précis, jamais des effets néon décoratifs.

**Probabilité :** 0,09

## Direction retenue — Atelier de contrôle local

### Design Movement

**Industrial modernism** appliqué à un outil de bureau : les matériaux visuels évoquent un poste de contrôle physique, mais l’interface conserve la finesse typographique et la clarté d’un logiciel de travail contemporain.

### Core Principles

1. La confiance vient d’abord de la lisibilité : chaque statut, action et limite est explicite.
2. La densité est utile, jamais compacte : les surfaces de travail s’organisent en bandes et en panneaux respirants.
3. Les signaux colorés sont réservés aux décisions opérationnelles ; l’interface normale reste minérale et calme.
4. Le local-first est rendu visible par des références aux machines, au coffre et aux chemins de données, sans imagery anxiogène.

### Color Philosophy

Le fond **basalte chaud** traduit le contrôle et l’ancrage local. Une encre bleue presque noire organise les surfaces, tandis que le **vert atelier** (`#B7F54A`) identifie l’action autorisée et les états sains. L’orange terre cuite n’apparaît que pour les interventions à vérifier ; le rouge reste réservé aux erreurs réelles. Cette palette évite les gradients violets et les conventions cyberpunk.

### Layout Paradigm

Le dashboard est conçu comme un **établi à trois rails** : une colonne de navigation étroite, une zone de travail principale et un rail d’observabilité pour les contrôles locaux. Les informations ne sont pas centrées dans une carte unique ; elles se lisent comme des plaques posées sur un plan de travail.

### Signature Elements

1. Un monogramme de forge abstrait, utilisé comme poinçon de marque et favicon.
2. Des repères de coordonnées et numéros de lots sur les panneaux de données.
3. Un motif de lignes topographiques très discret pour rappeler l’environnement local et les routes de profils.

### Interaction Philosophy

Les actions fréquentes répondent immédiatement avec un mouvement court et une confirmation lisible. Les actions sensibles restent distinctes, décrites et demandent une intention explicite. Les vues utilisent des transitions discrètes de panneau plutôt que des changements de page théâtraux.

### Animation

Les entrées de panneaux emploient uniquement l’opacité et une translation de 6 px sur 180 à 240 ms avec `cubic-bezier(0.23, 1, 0.32, 1)`. Les rangées se soulèvent légèrement au survol ; les boutons se compriment à `scale(0.97)` à l’activation. Les animations décoratives sont désactivées avec `prefers-reduced-motion`.

### Typography System

Les titres et chiffres opérationnels utilisent **Space Grotesk** pour leur construction mécanique. Les textes, tables et détails utilisent **IBM Plex Sans** pour la lecture prolongée. Les libellés de système et empreintes utilisent **IBM Plex Mono**, toujours en capitales espacées avec parcimonie.

### Brand Essence

**ForgeLocal est le poste de contrôle local pour organiser, isoler et auditer des profils navigateur sans dépendre d’un cloud imposé.**

Personnalité : **ancré, rigoureux, direct**.

### Brand Voice

Les titres annoncent un état ou une décision ; les CTA décrivent l’effet concret au lieu d’employer un vocabulaire promotionnel.

Exemples : « Vos profils restent sur cette machine. » et « Préparer un profil isolé ».

### Wordmark & Logo

Le mot-symbole est accompagné d’un poinçon abstrait : une forme de forge carrée ouverte, traversée par une ligne qui évoque à la fois un chemin local et un verrou. Le symbole fonctionne seul, sans texte, dans la barre latérale et le favicon.

### Signature Brand Color

**Vert atelier — `#B7F54A`**.

## Style Decisions

- Le vert atelier est réservé aux actions autorisées, états sains, sélection active et chiffres opérationnels majeurs ; les détails secondaires restent minéraux.
- Chaque panneau de données porte un repère d’instrumentation propriétaire : code de plaque, coordonnées ou poinçon de forge.
- Le monogramme ForgeLocal est traité comme un poinçon industriel récurrent, reconnaissable sans le mot-symbole.
