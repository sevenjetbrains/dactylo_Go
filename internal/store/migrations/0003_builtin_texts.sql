-- Pack de textes français du domaine public (F-25), pour s'entraîner sans
-- import. Inséré une seule fois : un texte supprimé ne revient pas.

-- +goose Up
-- +goose StatementBegin
INSERT INTO texts (title, content, language, source, tags) VALUES
('La Cigale et la Fourmi', 'La Cigale, ayant chanté
Tout l''été,
Se trouva fort dépourvue
Quand la bise fut venue :
Pas un seul petit morceau
De mouche ou de vermisseau.
Elle alla crier famine
Chez la Fourmi sa voisine,
La priant de lui prêter
Quelque grain pour subsister
Jusqu''à la saison nouvelle.
« Je vous paierai, lui dit-elle,
Avant l''Oût, foi d''animal,
Intérêt et principal. »
La Fourmi n''est pas prêteuse :
C''est là son moindre défaut.
« Que faisiez-vous au temps chaud ?
Dit-elle à cette emprunteuse.
— Nuit et jour à tout venant
Je chantais, ne vous déplaise.
— Vous chantiez ? j''en suis fort aise.
Eh bien ! dansez maintenant. »', 'fr', 'builtin', '["fable","La Fontaine"]'),

('Le Corbeau et le Renard', 'Maître Corbeau, sur un arbre perché,
Tenait en son bec un fromage.
Maître Renard, par l''odeur alléché,
Lui tint à peu près ce langage :
« Hé ! bonjour, Monsieur du Corbeau.
Que vous êtes joli ! que vous me semblez beau !
Sans mentir, si votre ramage
Se rapporte à votre plumage,
Vous êtes le Phénix des hôtes de ces bois. »
À ces mots le Corbeau ne se sent pas de joie ;
Et pour montrer sa belle voix,
Il ouvre un large bec, laisse tomber sa proie.
Le Renard s''en saisit, et dit : « Mon bon Monsieur,
Apprenez que tout flatteur
Vit aux dépens de celui qui l''écoute :
Cette leçon vaut bien un fromage, sans doute. »
Le Corbeau, honteux et confus,
Jura, mais un peu tard, qu''on ne l''y prendrait plus.', 'fr', 'builtin', '["fable","La Fontaine"]'),

('Demain, dès l''aube', 'Demain, dès l''aube, à l''heure où blanchit la campagne,
Je partirai. Vois-tu, je sais que tu m''attends.
J''irai par la forêt, j''irai par la montagne.
Je ne puis demeurer loin de toi plus longtemps.

Je marcherai les yeux fixés sur mes pensées,
Sans rien voir au dehors, sans entendre aucun bruit,
Seul, inconnu, le dos courbé, les mains croisées,
Triste, et le jour pour moi sera comme la nuit.

Je ne regarderai ni l''or du soir qui tombe,
Ni les voiles au loin descendant vers Harfleur,
Et quand j''arriverai, je mettrai sur ta tombe
Un bouquet de houx vert et de bruyère en fleur.', 'fr', 'builtin', '["poésie","Victor Hugo"]'),

('Le Dormeur du val', 'C''est un trou de verdure où chante une rivière,
Accrochant follement aux herbes des haillons
D''argent ; où le soleil, de la montagne fière,
Luit : c''est un petit val qui mousse de rayons.

Un soldat jeune, bouche ouverte, tête nue,
Et la nuque baignant dans le frais cresson bleu,
Dort ; il est étendu dans l''herbe, sous la nue,
Pâle dans son lit vert où la lumière pleut.

Les pieds dans les glaïeuls, il dort. Souriant comme
Sourirait un enfant malade, il fait un somme :
Nature, berce-le chaudement : il a froid.

Les parfums ne font pas frissonner sa narine ;
Il dort dans le soleil, la main sur sa poitrine
Tranquille. Il a deux trous rouges au côté droit.', 'fr', 'builtin', '["poésie","Rimbaud"]'),

('Chanson d''automne', 'Les sanglots longs
Des violons
De l''automne
Blessent mon cœur
D''une langueur
Monotone.

Tout suffocant
Et blême, quand
Sonne l''heure,
Je me souviens
Des jours anciens
Et je pleure ;

Et je m''en vais
Au vent mauvais
Qui m''emporte
Deçà, delà,
Pareil à la
Feuille morte.', 'fr', 'builtin', '["poésie","Verlaine"]'),

('L''Albatros', 'Souvent, pour s''amuser, les hommes d''équipage
Prennent des albatros, vastes oiseaux des mers,
Qui suivent, indolents compagnons de voyage,
Le navire glissant sur les gouffres amers.

À peine les ont-ils déposés sur les planches,
Que ces rois de l''azur, maladroits et honteux,
Laissent piteusement leurs grandes ailes blanches
Comme des avirons traîner à côté d''eux.

Ce voyageur ailé, comme il est gauche et veule !
Lui, naguère si beau, qu''il est comique et laid !
L''un agace son bec avec un brûle-gueule,
L''autre mime, en boitant, l''infirme qui volait !

Le Poète est semblable au prince des nuées
Qui hante la tempête et se rit de l''archer ;
Exilé sur le sol au milieu des huées,
Ses ailes de géant l''empêchent de marcher.', 'fr', 'builtin', '["poésie","Baudelaire"]'),

('Le Tour du monde en quatre-vingts jours (début)', 'En l''année 1872, la maison portant le numéro 7 de Saville-row, Burlington Gardens — maison dans laquelle Sheridan mourut en 1814 —, était habitée par Phileas Fogg, esq., l''un des membres les plus singuliers et les plus remarqués du Reform-Club de Londres, bien qu''il semblât prendre à tâche de ne rien faire qui pût attirer l''attention.

À l''un des plus grands orateurs qui honorent l''Angleterre, succédait donc ce Phileas Fogg, personnage énigmatique, dont on ne savait rien, sinon que c''était un fort galant homme et l''un des plus beaux gentlemen de la haute société anglaise.

On disait qu''il ressemblait à Byron — par la tête, car il était irréprochable quant aux pieds —, mais un Byron à moustaches et à favoris, un Byron impassible, qui aurait vécu mille ans sans vieillir.', 'fr', 'builtin', '["prose","Jules Verne"]'),

('Exercice : accents et majuscules', 'Portez ce vieux whisky au juge blond qui fume.

Voix ambiguë d''un cœur qui, au zéphyr, préfère les jattes de kiwis.

L''élève, ému, a reçu à Noël un cahier où il écrit déjà « Ça va très bien » en capitales : ÇA VA TRÈS BIEN.

Où êtes-vous, Agnès ? Là-bas, près de la forêt, sous un chêne âgé de cent ans.', 'fr', 'builtin', '["exercice","accents"]');
-- +goose StatementEnd

-- +goose Down
DELETE FROM texts WHERE source = 'builtin';
