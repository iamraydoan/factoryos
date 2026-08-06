# Business Domain: User Roles & Personas

## 1. Overview
Defines the user roles within the manufacturing facility, their daily responsibilities, and the pain points FactoryOS is designed to solve.

---

## 2. User Roles

### 👷 Operator
* **Daily Responsibilities:** Executes production operations at assigned Stations following Work Order instructions; scans material QR codes; reports defects; records output quantities.
* **Current Pain Points:** Relies on paper travelers for instructions; no digital guidance at the workstation; must find a supervisor when issues arise.
* **Needs from FactoryOS:** Step-by-step digital work instructions at the terminal; QR scan to verify correct materials; one-tap issue reporting.

### 👷‍♂️ Shift Leader
* **Daily Responsibilities:** Oversees a group of Operators within a single shift on a specific Line; resolves immediate floor-level issues; performs handover briefings between shifts.
* **Authority vs. Supervisor:** A Shift Leader has operational authority *within their shift only*. They can pause a station or reassign an Operator within their team, but **cannot** modify Work Order priority, alter routings, or approve Emergency WOs. Those actions require Production Supervisor approval.
* **Needs from FactoryOS:** Real-time Station status on a handheld device; shift handover summary report; ability to flag an issue for Supervisor escalation.

### 🏭 Production Supervisor
* **Daily Responsibilities:** Plans production shifts; assigns Operators and Shift Leaders to Lines; monitors Work Order progress across the full facility; resolves line stoppages and bottlenecks; approves Emergency Work Orders.
* **Current Pain Points:** No real-time visibility; must physically walk the floor or call Operators to check status.
* **Needs from FactoryOS:** Real-time dashboard for all Lines and Stations; automatic alerts when a Work Order is at risk of delay or a Station stops; ability to re-sequence WO priority.

### 📋 Production Planner
* **Daily Responsibilities:** Creates and manages the production schedule; resolves capacity conflicts between Work Orders; coordinates with ERP for material availability; communicates schedule changes to Supervisors.
* **Note:** This is a distinct role from the Supervisor. Planners own the **schedule**, Supervisors own the **floor execution**.
* **Needs from FactoryOS:** Finite capacity scheduling view; ability to drag-and-drop WO sequencing; at-risk WO alerts; integration hooks for ERP Production Orders.

### 🔍 Quality Inspector
* **Daily Responsibilities:** Performs inline quality inspections at production Stations; raises NCRs when defects are found; records disposition (Rework / Scrap / Use-As-Is); monitors CAPA resolution.
* **Current Pain Points:** Paper-based checklists; manual data re-entry; difficult to trace lot history across shifts.
* **Needs from FactoryOS:** Digital checklists at the workstation; one-tap NCR creation with automatic lot quarantine; instant lot genealogy lookup.

### 🔬 Quality Manager
* **Daily Responsibilities:** Owns the quality management system; reviews NCR trends; approves CAPA closure; interfaces with customers on quality escapes and recalls.
* **Authority vs. Inspector:** Quality Manager can approve Use-As-Is dispositions, close CAPAs, and initiate customer recall processes. Inspectors cannot.
* **Needs from FactoryOS:** NCR trend dashboards; SPC control chart monitoring; CAPA pipeline view; instant recall genealogy query.

### 🔧 Maintenance Technician
* **Daily Responsibilities:** Responds to machine faults; executes scheduled Preventive Maintenance (PM) tasks; records maintenance history.
* **Current Pain Points:** Receives work orders via phone or paper; no historical repair context; must search for technical documentation manually.
* **Needs from FactoryOS:** Push notifications for machine faults; access to full asset repair history; digital PM checklist.

### 📊 Plant Manager
* **Daily Responsibilities:** Reviews OEE, throughput, and quality reports; approves Emergency Work Orders; makes strategic decisions for the facility.
* **Current Pain Points:** Reports are manually compiled end-of-day or end-of-week; no real-time operational data.
* **Needs from FactoryOS:** Real-time OEE dashboard; automated shift reports; performance comparison by Line, Shift, and Product.

### 📦 Warehouse Operator
* **Daily Responsibilities:** Receives raw material lots from suppliers; prints and applies QR lot labels; stages materials to production line buffers; records finished goods into inventory.
* **Note:** Warehouse Operators are **direct users** of FactoryOS for the material receiving, lot labeling, and staging workflows. They use a simplified mobile-friendly interface focused on scan-and-confirm actions.
* **Needs from FactoryOS:** Mobile-friendly lot receiving screen; QR label printing; staging allocation confirmation; finished goods check-in.

### 🖥️ System Administrator
* **Daily Responsibilities:** Configures assets, user accounts, production routings, alert thresholds, and BOM structures.
* **Needs from FactoryOS:** Admin panel for managing Master Data (Assets, Routings, BOMs, Thresholds); RBAC user permission management.
