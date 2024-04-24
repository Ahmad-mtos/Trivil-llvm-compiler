\NeedsTeXFormat{LaTeX2e}
\ProvidesClass{trivilreport}[ Trivil language report class]

\LoadClass[a4paper]{article}
\usepackage[top=2cm, bottom=2cm, right=2cm]{geometry}

\RequirePackage{xltxtra}
\RequirePackage{etoolbox}

\RequirePackage{polyglossia}
\setmainlanguage{russian}
\setotherlanguage{english}
\setkeys{russian}{babelshorthands=true, indentfirst=false}

\setmainfont{Times New Roman}
\setromanfont{Times New Roman} 
\setsansfont{Arial} 
\setmonofont{Courier New} 

\newfontfamily{\cyrillicfont}{Times New Roman} 
\newfontfamily{\cyrillicfontrm}{Times New Roman}
\newfontfamily{\cyrillicfonttt}{Courier New}
\newfontfamily{\cyrillicfontsf}{Arial}

\addto\captionsrussian{%
  \renewcommand{\figurename}{Рис.}%
  \renewcommand{\tablename}{Табл.}%
}

%
% Commands for language name and company name
%
\newcommand{\thelang}{Тривиль}
\newcommand{\theauthor}{Алексей Недоря}

%=============

%\RequirePackage{iftex}

%\RequirePackage{unicode-math}
%\defaultfontfeatures{Scale=MatchLowercase}
%\defaultfontfeatures[\rmfamily]{Ligatures=TeX,Scale=1}

\RequirePackage{multirow}

\RequirePackage{xcolor}
\definecolor{darkblue}{rgb}{0.0, 0.0, 0.55}

% Index commands
\RequirePackage{imakeidx}
\makeindex[columns=3, title=Alphabetical Index, intoc]

%\RequirePackage{authblk}

\RequirePackage{fancyvrb}
\renewcommand{\theFancyVerbLine}{%
{\footnotesize\arabic{FancyVerbLine}}}

\DefineVerbatimEnvironment%
{Trivil}{Verbatim}
{frame=single,numbers=left}

\newcommand{\keyword}[1]{\textcolor{darkblue}{#1}}
\newcommand{\cautiously}[1]{\textcolor{red}{#1}}

\DefineVerbatimEnvironment%
{Grammar}{Verbatim}
{frame=lines,formatcom=\color{blue},rulecolor=\color{black}}

\DefineVerbatimEnvironment%
{SampleErr}{Verbatim}
{frame=single,numbers=left, rulecolor=\color{red}}

% Highlights
\RequirePackage{soul}
\newcommand{\needlink}{\sethlcolor{yellow}\hl{link}}
\newcommand{\TBD}{\sethlcolor{orange}\hl{TBD}}

\RequirePackage[numbered]{bookmark}

\makeatletter

% hyperref should be last import
\RequirePackage{hyperref}
\hypersetup{unicode=true, bookmarks=true}

% Internal refs
\newcommand{\See}[1]{\S\ref{#1}\seechecknextarg}
\newcommand{\seechecknextarg}{\@ifnextchar\bgroup{\seegobblenextarg}{}}
\newcommand{\seegobblenextarg}[1]{\@ifnextchar\bgroup{, \S\ref{#1}\seegobblenextarg}{ \textit{and} \S\ref{#1}}}

%External refs
\newcommand{\myref}[1]{\textcolor{blue}{#1}}

\makeatother

%dense itemize
\newenvironment{d_itemize}{
\begin{itemize}
  \setlength{\itemsep}{2pt}
  \setlength{\parskip}{0pt}
  \setlength{\parsep}{0pt}
}{\end{itemize}}

%==================

\hypersetup{
  pdftitle={\thelang{} Language Report},
  hidelinks,
  pdfcreator={\theauthor},
  pdfauthor={\theauthor},
  pdfsubject={\thelang{} language report},
  pdfkeywords={\theauthor{}, \thelang{}, language report}
}

