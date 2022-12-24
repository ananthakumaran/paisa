<nav class="navbar is-light" role="navigation" aria-label="main navigation">
    <div class="navbar-brand">
        <span class="navbar-item is-size-4 has-text-weight-medium">₹ Paisa</span>
        <a role="button" class="navbar-burger" aria-label="menu" aria-expanded="false" data-target="navbarBasicExample">
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
        </a>
    </div>

    <div class="navbar-menu">
        <div class="navbar-start">
            <a id="overview" class="navbar-item is-active">Overview</a>
            <a id="expense" class="navbar-item">Expense</a>
            <div class="navbar-item has-dropdown is-hoverable">
                <a class="navbar-link">Assets</a>
                <div class="navbar-dropdown">
                    <a id="investment" class="navbar-item">Investment</a>
                    <a id="gain" class="navbar-item">Gain</a>
                    <a id="allocation" class="navbar-item">Allocation</a>
                </div>
            </div>
            <a id="income" class="navbar-item">Income</a>
            <div class="navbar-item has-dropdown is-hoverable">
                <a class="navbar-link">Ledger</a>
                <div class="navbar-dropdown">
                    <a id="holding" class="navbar-item">Holding</a>
                    <a id="journal" class="navbar-item">Journal</a>
                    <a id="price" class="navbar-item">Price</a>
                </div>
            </div>
            <div class="navbar-item has-dropdown is-hoverable">
                <a class="navbar-link">Tax</a>
                <div class="navbar-dropdown">
                    <a id="harvest" class="navbar-item">Harvest</a>
                    <a id="schedule_al" class="navbar-item">Schedule AL</a>
                </div>
            </div>
            <a id="doctor" class="navbar-item">Doctor</a>
        </div>
    </div>
</nav>

<slot></slot>
